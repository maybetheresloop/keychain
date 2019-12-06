package main

//import (
//	"bufio"
//	"encoding/csv"
//	"flag"
//	"io"
//	"log"
//	"os"
//
//	"github.com/maybetheresloop/keychain/internal/data"
//)
//
//var fp string
//var out string
//
//const UsageFp = "CSV file to create database from"
//const UsageOut = "Database file to create"
//
//func init() {
//	flag.StringVar(&out, "o", "keychain.db", UsageOut)
//	flag.StringVar(&out, "out", "keychain.db", UsageOut)
//
//	flag.StringVar(&fp, "file", "", UsageFp)
//	flag.StringVar(&fp, "f", "", UsageFp)
//}
//
//func main() {
//	flag.Parse()
//
//	outFile, err := os.Create(out)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer outFile.Close()
//
//	var in io.Reader
//	if fp == "" {
//		in = os.Stdin
//	} else {
//		if in, err = os.Open(fp); err != nil {
//			log.Fatal(err)
//		}
//	}
//
//	r := csv.NewReader(bufio.NewReader(in))
//	w := data.NewWriter(outFile)
//
//	count := 0
//
//	for {
//		record, err := r.Read()
//		if err == io.EOF {
//			break
//		}
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		if len(record) < 2 {
//			log.Fatal("record must have at least two fields")
//		}
//
//		if err := w.WriteItem(data.NewItem([]byte(record[0]), []byte(record[1]))); err != nil {
//			log.Fatal(err)
//		}
//		count += 1
//	}
//
//	if err := w.Flush(); err != nil {
//		log.Fatal(err)
//	}
//
//	log.Printf("wrote %d records", count)
//}
