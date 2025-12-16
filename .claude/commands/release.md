Perform a release for dbsurf. The argument provided is: "$ARGUMENTS"

Steps to follow:

1. First, check if working directory is clean:
   - Run `git status --porcelain`
   - If there are uncommitted changes, stop and ask the user to commit or stash them

2. Fetch latest tags and branches:
   - Run `git fetch origin --tags`

3. Get the current latest tag:
   - Run `git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"`
   - Show the user the current version

4. Calculate the new version:
   - Parse the base version (strip any -alpha/-beta suffix)
   - Increment the patch number
   - If argument is "alpha", append "-alpha"
   - If argument is "beta", append "-beta"
   - If no argument, create a stable release (no suffix)
   - Show the user what the new version will be

5. Sync main with dev:
   - Run `git checkout main`
   - Run `git merge --ff-only dev`
   - Run `git push origin main`

6. Create and push the tag:
   - Run `git tag <new-version>`
   - Run `git push origin <new-version>`

7. Return to dev branch:
   - Run `git checkout dev`

8. Confirm success and show the GitHub Actions URL where they can monitor the release.
