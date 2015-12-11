#!/bin/bash

FAIL=1

while [ $FAIL -eq 1 ]; do
  go test -v ./... > auto_test.log
  grep FAIL auto_test.log
  FAIL=$?
done
