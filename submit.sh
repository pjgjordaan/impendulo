#! /bin/bash

git add . -A
git commit -m $0
git push --set-upstream origin master
git push --set-upstream backup master