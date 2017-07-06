#!/bin/bash
go run main.go iDiff 0cb40641836c e7d168d7db45 dir -j > tests/busybox_diff_actual.json
diff=$(diff tests/busybox_diff_expected.json tests/busybox_diff_actual.json)
if [ $diff ]; then
  echo "iDiff output is not as expected"
  exit 1
fi
