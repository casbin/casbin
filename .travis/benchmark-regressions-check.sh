#!/usr/bin/env bash

# visit https://docs.travis-ci.com/user/environment-variables/#default-environment-variables for details

version=$(go version | grep -Eo '[0-9]+\.[0-9]+')

# check environment
if [[ $TRAVIS_PULL_REQUEST == "false" ]] || [[ $version != "1.14" ]] || [[ $TRAVIS != "true" ]]; then
  echo "The script runs only on Travis-CI PR events with Golang 1.14"
  exit 0
fi

# install jq 1.6
curl -s -L https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux64 >./jq
chmod +x ./jq

# install benchstat
go get -u golang.org/x/perf/cmd/benchstat

remote_repo="https://github.com/$TRAVIS_REPO_SLUG"

echo "git clone --single-branch --branch ""$TRAVIS_BRANCH"" ""$remote_repo"" old_repo"
git clone --single-branch --branch "$TRAVIS_BRANCH" "$remote_repo" old_repo

cd old_repo || exit 1
go test -bench=. -count=3 | tee output.txt
cd - || exit 1

go test -bench=. -count=3 | tee output.txt

cat <<EOF >body.md
The computes and compares statistics about benchmarks for ${TRAVIS_PULL_REQUEST_SHA} by [benchstat](https://github.com/golang/perf/tree/master/cmd/benchstat) on Golang 1.14:

\`\`\`
$(benchstat old_repo/output.txt output.txt | tee benchstat.log)
\`\`\`
EOF

URL="https://api.github.com/repos/$TRAVIS_REPO_SLUG/issues/$TRAVIS_PULL_REQUEST/comments"

data='{ "body": '"$(./jq -R -s <body.md)"' }'

echo "request - $URL"

res=$(curl -s \
  -X POST \
  "$URL" \
  -H "Content-Type: application/json" \
  -H "Authorization: token $GITHUB_TOKEN" \
  --data "$data")

echo "$res"

url=$(echo "$res" | ./jq -r '.url')

if [[ -z $url || $url == null ]]; then
  exit 1
fi
