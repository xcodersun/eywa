#!/bin/bash

iters=10
verbose=0
log=auto_test.log
s=1
utilFail=0
integration=0
testname=""
tag=""
override=0

OPTIND=1


read -r -d '' USAGE << EOM
Usage:
  -h show this helper message.
  -? show this helper message.
  -f stop util fail. default: disabled
  -v enable verbose mode. default: disabled
  -n max iteration number, 0 means infinite. default: 10
  -s sleep seconds between iterations. default: 1 second
  -l log output. default: auto_test.log
  -i integration tests. default: disabled
  -t tests tag. default: none
  -T tests name. default: none
EOM

while getopts "h?voifn:s:l:t:T:" opt; do
    case "$opt" in
    h|\?)
        echo "$USAGE"
        exit 0
        ;;
    v)  verbose=1
        ;;
    f)  utilFail=1
        ;;
    i)  integration=1
        tag=integration
        ;;
    l)  log=$OPTARG
        ;;
    s)  s=$OPTARG
        ;;
    n)  iters=$OPTARG
        ;;
    t)  tag=$OPTARG
        ;;
    T)  testname=$OPTARG
        ;;
    o)  override=1
        ;;
    esac
done

shift $((OPTIND-1))

if [[ $utilFail -eq 1 ]]; then
  verbose=1
fi

if [[ $iters -eq 0 && $utilFail -eq 0 ]]; then
  echo "WARN: Test will run infinitely. Are you sure? [Y/N]:"
  read c
  if [ "$c" != "Y" ]; then
    exit 0;
  fi
fi

params="$([[ $verbose != 1 ]] || echo '-v' ) $([[ ${#tag} -gt 0 ]] && echo '-tags' ${tag}) $([[ ${#testname} -gt 0 ]] && echo '--run' ${testname}) ./..."

# echo "go test $params >> $log 2>&1"

echo "Starting Tests..."

serverpid=""

if [[ $integration -eq 1 ]]; then
  echo "Starting test server..."
  cd tasks; go run migration.go -conf=$PWD/../configs/octopus_test.yml; cd ..
  ./octopus -conf=$PWD/configs/octopus_test.yml &
  serverpid=$!
  sleep 3
fi

fail=1
iter=0
rm $log >/dev/null 2>&1

while [[ $utilFail -eq 1 && $iters -eq 0 && $fail -eq 1 ]] || [[ $utilFail -eq 1 && $iters -gt 0 && $fail -eq 1 && $iter -lt $iters ]] || [[ $utilFail -ne 1 && $iters -eq 0 ]] || [[ $utilFail -ne 1 && $iters -gt 0 && $iter -lt $iters ]]; do

  if [[ $override -eq 1 ]]; then
    rm $log >/dev/null 2>&1
  fi

  echo "TEST RUN [$iter] ========================================================================" >> $log

  go test $params >> $log 2>&1

  grep FAIL $log
  fail=$?
  iter=$((iter+1))
  sleep $s
done

echo "Tests are done..."

echo "Test results are in $log. iterations=$iter"

if [[ ${#serverpid} -gt 0 ]]; then
  echo "Shutting down test server..."

  kill -INT $serverpid

  ps auxww | grep $serverpid | grep -v grep
  running=$?

  while [[ $running -eq 0 ]]; do
    sleep 1

    ps auxww | grep $serverpid | grep -v grep
    running=$?
  done

  echo "Test server is shutdown..."
fi
