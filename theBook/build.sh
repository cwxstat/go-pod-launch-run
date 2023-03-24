#!/bin/bash
mdbook build
rm -rf ../docs
mv book ../docs
cp CNAME ../docs
git add ../docs
mdbook clean
git add ../theBook
git commit -m "update book"
git push
