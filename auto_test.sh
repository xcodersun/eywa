#!/bin/bash

iters=1
verbose=0
log=auto_test.log
s=1
utilFail=0
integration=0
testname=""
tag=""
override=0
timeout=0
rebuild=0

OPTIND=1

if [[ "${#EYWA_HOME}" -eq 0 ]]; then
  echo 'Warn: EYWA_HOME is not set. Using current directory as EYWA_HOME.'
  export EYWA_HOME=$PWD
fi

read -r -d '' USAGE << EOM
Usage:
  -h show this helper message.
  -? show this helper message.
  -f stop util fail. default: disabled
  -v enable verbose mode. default: disabled
  -n max iteration number, 0 means infinite. default: 1
  -s sleep seconds between iterations. default: 1 second
  -l log output. default: auto_test.log
  -i integration tests. default: disabled
  -t tests tag. default: none
  -T tests name. default: none
  -o override the log in each of the test iteration
  -b rebuild binary before running integration tests. default: not rebuild
EOM

while getopts "h?vobifn:s:l:t:T:O:" opt; do
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
    O)  timeout=$OPTARG
	;;
    b)  rebuild=1
        ;;
    esac
done

shift $((OPTIND-1))

if [[ $utilFail -eq 1 ]]; then
  verbose=1
fi

if [[ $iters -eq 0 && $utilFail -eq 1 ]]; then
  echo "WARN: Test will run infinitely. Are you sure? [Y/N]:"
  read c
  if [ "$c" != "Y" ]; then
    exit 0;
  fi
fi

params="$([[ $verbose != 1 ]] || echo '-v' ) \
$([[ $timeout -gt 0 ]] && echo '-timeout' ${timeout}s) \
$([[ ${#tag} -gt 0 ]] && echo '-tags' ${tag}) \
$([[ ${#testname} -gt 0 ]] && echo '--run' \
${testname}) ./..."

serverpid=""

if [[ $integration -eq 1 ]]; then
  if [[ $rebuild -eq 1 ]]; then
    echo "Rebuilding eywa binary..."
    go build -a
  fi

  echo "Running test setup..."
  "$EYWA_HOME"/eywa migrate -conf="$EYWA_HOME/configs/eywa_test.yml"
  curl -XDELETE 127.0.0.1:9200/channels.* > /dev/null 2>&1
  curl -XDELETE 127.0.0.1:9200/_template/* > /dev/null 2>&1
  "$EYWA_HOME"/eywa setup_es -conf="$EYWA_HOME/configs/eywa_test.yml"

  echo "Starting test server..."
  "$EYWA_HOME"/eywa serve -conf="$EYWA_HOME/configs/eywa_test.yml" &
  sleep 5
  serverpid="$(cat "${EYWA_HOME}"/tmp/pids/eywa_test.pid)"
  if [[ ${#serverpid} -eq 0 ]]; then
    echo 'Error: test server was not started within 5 seconds'
    exit 1
  fi
  echo "Test server started..."
fi

fail=1
iter=0
echo '' > $log

echo "Starting Tests..."

while [[ $utilFail -eq 1 && $iters -eq 0 && $fail -eq 1 ]] || [[ $utilFail -eq 1 && $iters -gt 0 && $fail -eq 1 && $iter -lt $iters ]] || [[ $utilFail -ne 1 && $iters -eq 0 ]] || [[ $utilFail -ne 1 && $iters -gt 0 && $iter -lt $iters ]]; do

  if [[ $override -eq 1 ]]; then
    echo '' > $log
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

  ps auxww | grep $serverpid | grep -v grep >/dev/null 2>&1
  running=$?

  waitIter=0
  while [[ $running -eq 0 ]]; do
    echo "Trying ${waitIter} time."
    if [[ $waitIter -gt 10 ]]; then
      echo 'Error: server failed to shutdown.'
      exit 1
    fi

    sleep 1

    ps auxww | grep $serverpid | grep -v grep >/dev/null 2>&1
    running=$?
    waitIter=$((waitIter+1))
  done

  echo "Test server is shutdown..."
fi

if [[ $fail -ne 1 ]]; then
  exit 1
fi
