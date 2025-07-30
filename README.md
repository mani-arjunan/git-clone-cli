# Git clone CLI tool

This go code lets u too clone ur repository under any of ur organisations that you are part of.

## Prerequisites

- You need to create your github personal access token under the settings, full details [here](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens)
- Once done have it under an env variable called GH_TOKEN.
- Also mention ur directory path on where u have to clone CLONE_DIR, if not provided the default path would be `user_home/Documents`

## Run the app

- Once all the above prerequisites are done you are good to go, now run `go run main.go`.
- In MAC you can also install with brew `brew tap mani-arjunan/gic && brew install gic`, for any updates do `brew update && brew reinstall gic`.
- In linux based distros, extract the tar under releases page(use latest).

I have created this script to not forget my golang knowledge, 
coz it's been very very long time on working golang professionally, am stuck with this mf js rn,
nthg special engineering or any blazingly fast algorithms are used, just a plain 
api call wrapper to the stdout.

> **_NOTE:_**  Created fully by mani! not by any gpt, claude or grok dumbass's, with this in mind code may have O(n^2) of bugs, <br/>
> Feel free to roast it, review it, find bugs etccc




