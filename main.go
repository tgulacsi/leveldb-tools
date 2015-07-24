// Copyright 2015 Tamás Gulácsi
//
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

var sep = [...]byte{'-', '>'}

func main() {
	rootCmd := &cobra.Command{
		Short: "leveldb-tools",
	}
	var db *leveldb.DB
	openDB := func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			rootCmd.Usage()
			os.Exit(1)
		}
		var err error
		opts := &opt.Options{ErrorIfMissing: cmd.Use == "dump", Strict: opt.StrictAll}
		if db, err = leveldb.OpenFile(args[0], opts); err != nil {
			log.Fatal(err)
		}
	}

	dumpCmd := &cobra.Command{
		Use:   "dump",
		Short: "dump database",
		Run: func(cmd *cobra.Command, args []string) {
			defer db.Close()
			defer os.Stdout.Close()
			w := bufio.NewWriter(os.Stdout)
			defer w.Flush()
			iter := db.NewIterator(nil, &opt.ReadOptions{Strict: opt.StrictAll, DontFillCache: true})
			defer iter.Release()
			sep := sep[:]
			for iter.Next() {
				k, v := iter.Key(), iter.Value()
				w.WriteString(fmt.Sprintf("+%d,%d:", len(k), len(v)))
				w.Write(k)
				w.Write(sep)
				w.Write(v)
				if err := w.WriteByte('\n'); err != nil {
					log.Fatal(err)
				}
			}
			if err := iter.Error(); err != nil {
				log.Fatal(err)
			}
		},
		PersistentPreRun: openDB,
	}
	rootCmd.AddCommand(dumpCmd)

	loadCmd := &cobra.Command{
		Use:   "load",
		Short: "load database",
		Run: func(cmd *cobra.Command, args []string) {
			defer func() {
				if err := db.Close(); err != nil {
					log.Fatal(err)
				}
			}()
			r := bufio.NewReader(os.Stdin)
			var lk, lv int
			var k, v []byte
			sepLen := len(sep)
			sep := sep[:]
			for {
				if _, err := fmt.Fscanf(r, "+%d,%d:", &lk, &lv); err != nil {
					if err == io.EOF || err == io.ErrUnexpectedEOF {
						break
					}
					log.Fatal(err)
				}
				log.Printf("lk=%d, lv=%d", lk, lv)
				if cap(k) < lk {
					k = make([]byte, lk*2)
				}
				if cap(v) < lv+sepLen+1 {
					v = make([]byte, lv*2+2)
				}
				n, err := io.ReadFull(r, k[:lk])
				if err != nil {
					log.Fatal(err)
				}
				k = k[:n]
				if n, err = io.ReadFull(r, v[:lv+sepLen+1]); err != nil {
					log.Fatal(err)
				}
				if !bytes.Equal(sep, v[:sepLen]) {
					log.Fatal("awaited %q, got %q", sep, v)
				}
				if v[n-1] != '\n' {
					log.Fatal("should end with EOL, got %q", v)
				}
				v = v[2 : n-1]
				if err := db.Put(k, v, nil); err != nil {
					log.Fatal(err)
				}
			}
		},
		PersistentPreRun: openDB,
	}
	rootCmd.AddCommand(loadCmd)
	rootCmd.Execute()
}
