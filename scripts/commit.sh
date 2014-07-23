#! /bin/bash
(cd $IMPENDULO && ls && git add . -A && git commit -m "$1" && git push origin $2 && git push backup $2)
