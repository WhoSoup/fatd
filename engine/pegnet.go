package engine

import (
	"fmt"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/fatd/db"
	"github.com/Factom-Asset-Tokens/fatd/fat"
	"github.com/Factom-Asset-Tokens/fatd/flag"
)

func PegNetOpenOrCreate(dbpath string, id *factom.Bytes32) (Chain, error) {
	fname := id.String() + "/opr.sqlite3"
	var chain Chain
	dbchain, err := db.OpenPegNet(dbpath, fname)
	if err != nil {
		err = createDir(dbpath + id.String())
		if err != nil {
			return chain, err
		}
		err = chain.OpenNewPegNetbyID(c, id)
		if err != nil {
			return chain, err
		}

	} else {
		chain.Chain = dbchain
	}

	return chain, nil
}

func (chain *Chain) OpenNewPegNetbyID(c *factom.Client, chainID *factom.Bytes32) error {
	eblocks, err := factom.EBlock{ChainID: chainID}.GetPrevAll(c)
	if err != nil {
		return fmt.Errorf("factom.EBlock{}.GetPrevAll(): %v", err)
	}

	first := eblocks[len(eblocks)-1]
	// Get DBlock Timestamp and KeyMR
	var dblock factom.DBlock
	dblock.Header.Height = first.Height
	if err := dblock.Get(c); err != nil {
		return fmt.Errorf("factom.DBlock{}.Get(): %v", err)
	}
	first.SetTimestamp(dblock.Header.Timestamp)
	*chain, err = OpenNewPegNet(c, dblock.KeyMR, first)
	if err != nil {
		return err
	}
	if chain.IsUnknown() {
		return fmt.Errorf("not a valid FAT chain: %v", chainID)
	}

	// We already applied the first EBlock. Sync the remaining.
	return chain.SyncEBlocks(c, eblocks[:len(eblocks)-1])
}

func OpenNewPegNet(c *factom.Client,
	dbKeyMR *factom.Bytes32, eb factom.EBlock) (chain Chain, err error) {
	if err := eb.Get(c); err != nil {
		return chain, fmt.Errorf("%#v.Get(c): %v", eb, err)
	}
	// Load first entry of new chain.
	first := &eb.Entries[0]
	if err := first.Get(c); err != nil {
		return chain, fmt.Errorf("%#v.Get(c): %v", first, err)
	}
	if !eb.IsFirst() {
		return
	}

	// Ignore chains with NameIDs that don't match the fat pattern.
	nameIDs := first.ExtIDs
	if !fat.ValidPegNetOracleIDs(nameIDs, "OraclePriceRecords") {
		return
	}

	var identity factom.Identity
	identity.ChainID = factom.NewBytes32(nil)

	if err := eb.GetEntries(c); err != nil {
		return chain, fmt.Errorf("%#v.GetEntries(c): %v", eb, err)
	}

	chain.Chain, err = db.NewPegNet(flag.DBPath, dbKeyMR, eb, flag.NetworkID, identity)
	if err != nil {
		return chain, fmt.Errorf("db.OpenNew(): %v", err)
	}
	chain.ChainStatus = ChainStatusTracked
	chain.Pending.Chain = chain.Chain
	return
}
