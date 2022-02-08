#! /bin/bash
set -e
cd "${0%/*}"

ENV=$1
PROJECT=$2

if [ "$ENV" != dev ] && [ "$ENV" != prod ]; then
	echo -e "\n\033[41mUSAGE: $0 {prod | dev}\033[0m\n"
    exit 1
fi

if [ -z "$PROJECT" ]
then
      echo -e "\n\033[41mUSAGE: $0 PROJECT variable not set\033[0m\n"
fi

echo -e "\033[33m\nInitiating release for PlanB:\033[0m\n"

if [ -n "$(git status --porcelain)" ]; then
  echo -e '\n\033[31mEnsure there are no changes on LOCAL to make a release. Aborting script.\033[0m';
  exit 1;
fi

BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [ "$ENV" = "prod" ]; then
    if [[ "$BRANCH" != "master" ]]; then
        echo -e '\n\033[31mYou need to be on MASTER to release to production. Aborting script.\033[0m';
        exit 1;
    fi
fi

REMOTE_HEAD=$(git rev-parse origin/$BRANCH)
LOCAL_HEAD=$(git rev-parse HEAD)
if [ "$REMOTE_HEAD" != "$LOCAL_HEAD" ]; then
  echo -e '\n\033[31mREMOTE and LOCAL need to be up-to-date to make a release. Aborting script.\033[0m';
  exit 1;
fi

LATEST_TAG=$(git ls-remote --tags --sort="v:refname" git@github.com:vinayb21/planb.git | tail -n1 | sed 's/.*\///; s/\^{}//'| sed 's/^v//g')
LATEST_MAIN_VERSION=$(echo $LATEST_TAG | cut -d - -f 1)
if [[ "$LATEST_TAG" == *"-dev-"* ]]; then
    LATEST_DEV_VERSION=$(echo $LATEST_TAG | cut -d - -f 3)
else
    LATEST_DEV_VERSION=''
fi

if [ -z "$LATEST_MAIN_VERSION" ]; then
	NEXT_MAIN_VERSION=1
else
  # https://stackoverflow.com/questions/8653126/how-to-increment-version-number-in-a-shell-script
  NEXT_MAIN_VERSION=$(echo $LATEST_MAIN_VERSION | awk -F. -v OFS=. 'NF==1{print ++$NF}; NF>1{if(length($NF+1)>length($NF))$(NF-1)++; $NF=sprintf("%0*d", length($NF), ($NF+1)%(10^length($NF))); print}')
fi

if [ "$ENV" = "dev" ]; then
    if [ -z "$LATEST_DEV_VERSION" ]; then
        TAG=v$NEXT_MAIN_VERSION-dev-1
    else
        NEXT_DEV_VERSION=$(echo $LATEST_DEV_VERSION | awk -F. -v OFS=. 'NF==1{print ++$NF}; NF>1{if(length($NF+1)>length($NF))$(NF-1)++; $NF=sprintf("%0*d", length($NF), ($NF+1)%(10^length($NF))); print}')
        TAG=v$LATEST_MAIN_VERSION-dev-$NEXT_DEV_VERSION
    fi
else
    if [ -z "$LATEST_DEV_VERSION" ]; then
        TAG=v$NEXT_MAIN_VERSION
    else
        TAG=v$LATEST_MAIN_VERSION
    fi
fi

git tag -a $TAG -m ""
git push origin $TAG

cd ..
docker build -t gcr.io/$PROJECT/planb:$TAG .
docker push gcr.io/$PROJECT/planb:$TAG
