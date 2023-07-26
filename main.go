package main

import (
	"fmt"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"os"
	"path"
)

const (
	testRepo = "testRepo"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	stat, err := os.Stat(testRepo)
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}
	if stat != nil && stat.IsDir() {
		checkErr(os.RemoveAll(testRepo))
	}

	repo, err := git.PlainInit(testRepo, false)
	checkErr(err)

	worktree, err := repo.Worktree()
	checkErr(err)

	checkErr(createOrAppendToRepoFile("testfile01.txt", "testFile01"))
	checkErr(createOrAppendToRepoFile("testfile02.txt", "testFile02"))
	checkErr(createOrAppendToRepoFile("testfile03.txt", "testFile03"))
	checkErr(createOrAppendToRepoFile(".gitignore", "testfile03.txt\n"))
	checkErr(addRepoFile(worktree, "testfile01.txt"))
	checkErr(addRepoFile(worktree, "testfile02.txt"))
	checkErr(addRepoFile(worktree, "testfile03.txt"))
	checkErr(addRepoFile(worktree, ".gitignore"))
	_, err = worktree.Commit("first commit", &git.CommitOptions{
		Author:    &object.Signature{Name: "TestUser", Email: "testuser@example.com"},
		Committer: nil,
	})
	checkErr(err)

	checkErr(createOrAppendToRepoFile("testfile04.txt", "testFile04"))
	checkErr(deleteRepoFile("testfile03.txt"))

	// worktree is not clean!!!
	// Changes not staged for commit:
	//   deleted:    testfile03.txt
	// Untracked files:
	//   testfile04.txt

	// should be the same as 'git reset --hard'
	err = worktree.Reset(&git.ResetOptions{Mode: git.HardReset})
	checkErr(err)

	status, err := worktree.Status()
	checkErr(err)

	if status.IsClean() {
		fmt.Println("Everything is fine!")
	} else {
		fmt.Println("\n!!! ooops, wortree is still not clean after reset hard !!!")
		for file, fileStatus := range status {
			fmt.Printf("  %s: %s\n", string(fileStatus.Worktree), file)
		}
		fmt.Println()
		fmt.Println("This happens if the the modified file (in this case the deleted file testfile03.txt) is also included in .gitignore")
		fmt.Println("'git reset --hard' reverts the deletion in this case too!!!")
		fmt.Println("This behaviour is since https://github.com/go-git/go-git/commit/cf51e2febf37332f11ae63feca768d9672e10a36 (Change in worktree_status.go)")
	}
}

func addRepoFile(wt *git.Worktree, filename string) error {
	_, err := wt.Add(filename)
	return err
}

func createOrAppendToRepoFile(filename, content string) error {
	repoFilename := path.Join(testRepo, filename)
	var err error
	var file *os.File
	_, err = os.Stat(repoFilename)
	if os.IsNotExist(err) {
		file, err = os.Create(repoFilename)
		if err != nil {
			return err
		}
	} else {
		file, err = os.Open(repoFilename)
		if err != nil {
			return err
		}
	}

	_, err = file.WriteString(content)
	return err
}

func deleteRepoFile(filename string) error {
	return os.Remove(path.Join(testRepo, filename))
}
