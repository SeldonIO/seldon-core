## Undo Last Commit and Remove Unwanted Files from Staging

You accidentally committed `.DS_Store` files and want to undo the commit without pushing, and remove those files from staging.

### Terminal Session

```sh
➜  seldon-core git:(v2) ✗ git add doc/
➜  seldon-core git:(v2) ✗ git commit -m "remove references to core 1"
[v2 98400011d] remove references to core 1
 Committer: Rakavitha Kodhandapani <seldon@SELIN002.local>
Your name and email address were configured automatically based
on your username and hostname. Please check that they are accurate.
You can suppress this message by setting them explicitly:

    git config --global user.name "Your Name"
    git config --global user.email you@example.com

After doing this, you may fix the identity used for this commit with:

    git commit --amend --reset-author

 2 files changed, 0 insertions(+), 0 deletions(-)
 create mode 100644 doc/.DS_Store
 create mode 100644 doc/source/.DS_Store
```

### How to Fix

1. **Undo the last commit but keep changes in your working directory:**

    ```sh
    git reset HEAD~1
    ```

2. **Remove the unwanted files from staging:**

    ```sh
    git rm --cached doc/.DS_Store doc/source/.DS_Store
    ```

3. **(Optional) Add `.DS_Store` to your `.gitignore` to prevent future issues:**

    ```sh
    echo ".DS_Store" >> .gitignore
    git add .gitignore
    ```

4. **Re-commit your changes (if needed):**

    ```sh
    git commit -m "remove references to core 1"
    ```

---

**Note:**  
`.DS_Store` files are macOS system files and are usually not needed in repositories. Adding them to `.gitignore` is a best practice. 