# Life [![Build Status](https://travis-ci.org/vrecan/death.svg?branch=master)](https://travis-ci.org/vrecan/death)

Simple wrapper for handling creation and management of a single background goroutine.

# Why?

In most of our go programs we had a lot of boilerplate code that if done wrong would generate bugs.

common mistakes this helps us avoid:

- sync.Waitgroup without a pointer
- done channel with a size above 0 //if 0 and close is called but start isn't you block forever
- start is always in a Once.Do so repeated calls to start don't spin up multiple goroutines

# Release Process

## Rules for release branches:

- If you are releasing a new major version you need to branch off of master into a branch `release-branch.v#` (example `release-branch.v2` for a 2.x release)
- If you are releasing a minor or patch update to an existing major release make sure to merge master into the release branch

## Rules for tagging and publishing the release

When you are ready to publish the release make sure you...

1. Merge your changes into the correct release branch.
2. Check out the release branch locally (example: `git pull origin release-branch.v3`)
3. Create a new tag for the specific release version you will publish (example: `git tag v3.0.1`)
4. Push the tag up to github (example: `git push origin v3.0.1`)
5. Go to the release tab in github
6. Select the target branch as the release branch and type in the tag name (tagname should include `v` so example: `v3.0.1`)
7. Write a title and a well worded description on exactly what is in this change
8. Click publish release
