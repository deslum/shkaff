#!/bin/bash
git clone https://github.com/google/leveldb.git
cd leveldb/
make all
cd ..
rm -R leveldb
