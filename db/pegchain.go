// MIT License
//
// Copyright 2018 Canonical Ledgers, LLC
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

package db

import (
	"fmt"
	"os"
	"strings"

	"github.com/Factom-Asset-Tokens/fatd/db/pegnet"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/fatd/db/addresses"
	"github.com/Factom-Asset-Tokens/fatd/db/eblocks"
	"github.com/Factom-Asset-Tokens/fatd/db/entries"
	"github.com/Factom-Asset-Tokens/fatd/db/metadata"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/Factom-Asset-Tokens/fatd/fat/fat2"
	_log "github.com/Factom-Asset-Tokens/fatd/log"
	"github.com/pegnet/pegnet/modules/grader"
)

func OpenPegNet(dbPath, fname string) (chain Chain, err error) {
	chain.Conn, chain.Pool, err = OpenConnPool(dbPath + fname)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			chain.Close()
		}
	}()
	chain.Log = _log.New("chain", strings.TrimRight(fname, dbFileExtension))
	chain.DBFile = fname

	err = chain.loadPNMetadata()
	return
}

func NewPegNet(dbPath string,
	dbKeyMR *factom.Bytes32, eb factom.EBlock, networkID factom.NetworkID,
	identity factom.Identity) (chain Chain, err error) {
	fname := eb.ChainID.String() + "/opr" + dbFileExtension
	path := dbPath + fname

	nameIDs := eb.Entries[0].ExtIDs
	if !fat.ValidPegNetOracleIDs(nameIDs, "OraclePriceRecords") {
		err = fmt.Errorf("invalid token chain Name IDs")
		return
	}

	// Ensure that the database file doesn't already exist.
	_, err = os.Stat(path)
	if err == nil {
		err = fmt.Errorf("already exists: %v", path)
		return
	}
	if !os.IsNotExist(err) { // Any other error is unexpected.
		return
	}

	chain.Conn, chain.Pool, err = OpenConnPool(dbPath + fname)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			chain.Close()
			if err := os.Remove(path); err != nil {
				chain.Log.Errorf("os.Remove(): %v", err)
			}
		}
	}()
	chain.Log = _log.New("chain", strings.TrimRight(fname, dbFileExtension))
	chain.DBFile = fname
	chain.ID = eb.ChainID
	chain.DBKeyMR = dbKeyMR
	chain.SyncHeight = eb.Height
	chain.SyncDBKeyMR = dbKeyMR
	chain.NetworkID = networkID
	chain.Identity = identity
	chain.Type = fat.TypeFAT2

	if err = metadata.Insert(chain.Conn, chain.SyncHeight, chain.SyncDBKeyMR,
		chain.NetworkID, chain.Identity); err != nil {
		return
	}

	// Ensure that the coinbase address has rowid = 1.
	coinbase := fat.Coinbase()
	if _, err = addresses.Add(chain.Conn, &coinbase, 0); err != nil {
		return
	}

	chain.setPNApplyFunc()

	if err = chain.Apply(dbKeyMR, eb); err != nil {
		return
	}

	return
}

func (chain *Chain) setPNApplyFunc() {
	chain.apply = func(chain *Chain, ei int64, e factom.Entry) (
		txErr, err error) {
		_, txErr, err = chain.ApplyPNOracle(ei, e)
		return
	}
	chain.preEBlock = func(chain *Chain, keyMR *factom.Bytes32, eb factom.EBlock) (err error) {
		err = chain.PNPreEBlock(keyMR, eb)
		return
	}
	chain.postEBlock = func(chain *Chain, keyMR *factom.Bytes32, eb factom.EBlock) (err error) {
		err = chain.PNPostEBlock(keyMR, eb)
		return
	}
}

var tmp grader.BlockGrader

func (chain *Chain) PNPreEBlock(keyMR *factom.Bytes32, eb factom.EBlock) error {
	grader.InitLX()
	ver := uint8(1)
	fmt.Println("pre-eblock", eb.Sequence, eb.Height)
	if eb.Height >= 206422 {
		ver = uint8(2)
	}

	prev, err := pegnet.GetGrade(chain.Conn, eb.Height-1)
	if err != nil {
		return err
	}

	tmp, err = grader.NewGrader(ver, int32(eb.Height), prev)
	if err != nil {
		return err
	}
	return nil
}

func (chain *Chain) PNPostEBlock(keyMR *factom.Bytes32, eb factom.EBlock) error {
	fmt.Println("post e-block", tmp.Count())
	graded := tmp.Grade()

	fmt.Println(graded.WinnersShortHashes())
	//panic("")
	return nil
}

func (chain *Chain) ApplyPNOracle(ei int64, e factom.Entry) (tx *fat2.OraclePriceRecord,
	txErr, err error) {

	if tmp == nil {
		return
	}
	var extids [][]byte
	for _, x := range e.ExtIDs {
		extids = append(extids, []byte(x))
	}

	err = tmp.AddOPR(e.Hash[:], extids, []byte(e.Content))
	if err != nil {
		fmt.Println(err)
	}
	return
}

func (chain *Chain) loadPNMetadata() error {
	defer chain.setPNApplyFunc()
	// Load NameIDs
	first, err := entries.SelectByID(chain.Conn, 1)
	if err != nil {
		return err
	}
	if !first.IsPopulated() {
		return fmt.Errorf("no first entry")
	}

	nameIDs := first.ExtIDs
	if !fat.ValidPegNetOracleIDs(nameIDs, "OraclePriceRecords") {
		return fmt.Errorf("invalid token chain Name IDs")
	}

	// Load Chain Head
	eb, dbKeyMR, err := eblocks.SelectLatest(chain.Conn)
	if err != nil {
		return err
	}
	if !eb.IsPopulated() {
		// A database must always have at least one EBlock.
		return fmt.Errorf("no eblock in database")
	}
	chain.Head = eb
	chain.DBKeyMR = &dbKeyMR
	chain.ID = eb.ChainID

	chain.SyncHeight, chain.NumIssued, chain.SyncDBKeyMR,
		chain.NetworkID, chain.Identity,
		chain.Issuance, err = metadata.Select(chain.Conn)
	chain.Type = fat.TypeFAT2
	return err
}
