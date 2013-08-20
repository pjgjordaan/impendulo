#! /bin/bash
git add . -A
git commit -m "$1"
git push --set-upstream origin master
git push --set-upstream backup master
