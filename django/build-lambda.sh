#!/usr/bin/env bash

set -eux
zip -r9 package.zip .  -x ".git/*" -x "terraform/*" -x "package/*"
poetry export -f requirements.txt -o requirements.txt
pip install --target=./package -r requirements.txt
cd package
zip -r9 -g ../package.zip *
cd ..
