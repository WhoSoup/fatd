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

// Package nftokens provides functions and SQL framents for working with the
// "nf_tokens" table, which stores fat.NFToken with owner, creation id, and
// metadata.

package pegnet

import (
	"encoding/json"
	"fmt"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"github.com/Factom-Asset-Tokens/factom"
)

const CreateTableGrade = `CREATE TABLE "pn_grade" (
        "eb_seq" INTEGER NOT NULL,
        "winners" BLOB,
        
        UNIQUE("eb_seq")

        FOREIGN KEY("eb_seq") REFERENCES "eblocks"
);
`

func InsertGrade(conn *sqlite.Conn, eb factom.EBlock, winners []string) (error, error) {
	data, err := json.Marshal(winners)
	if err != nil {
		return nil, err
	}

	stmt := conn.Prep(`INSERT INTO "pn_grade"
                ("eb_seq", "winners") VALUES (?, ?);`)
	stmt.BindInt64(1, int64(eb.Sequence))
	stmt.BindBytes(2, data)
	if _, err := stmt.Step(); err != nil {
		if sqlite.ErrCode(err) == sqlite.SQLITE_CONSTRAINT_UNIQUE {
			return fmt.Errorf("Grade{%d} already exists", eb.Sequence), nil
		}
		return nil, err
	}
	return nil, nil
}

func GetGrade(conn *sqlite.Conn, id uint32) ([]string, error) {
	stmt := conn.Prep(`SELECT "winners" FROM "pn_grade" WHERE "eb_seq" = ?;`)
	stmt.BindInt64(1, int64(id))
	raw, err := sqlitex.ResultText(stmt)
	if err != nil && err.Error() == "sqlite: statement has no results" {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var winners []string
	err = json.Unmarshal([]byte(raw), &winners)
	if err != nil {
		return nil, err
	}

	return winners, nil
}
