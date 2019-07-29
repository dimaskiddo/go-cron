REBASE_URL         := "github.com/dimaskiddo/go-cron"
COMMIT_MSG         := "update improvement"

.PHONY:

.SILENT:

init:
	make clean
	dep init -v

init-dist:
	mkdir -p dist
	touch dist/.gitkeep

ensure:
	make clean
	dep ensure -v

release:
	make ensure
	goreleaser --snapshot --skip-publish --rm-dist
	make init-dist
	echo "Release complete please check dist directory."

publish:
	GITHUB_TOKEN=$(GITHUB_TOKEN) goreleaser --rm-dist
	make clean-dist
	echo "Publish complete please check your repository releases."

run:
	go run *.go

clean-dist:
	rm -rf ./dist/*
	make init-dist

clean:
	make clean-dist
	rm -rf ./vendor

commit:
	make ensure
	make clean
	git add .
	git commit -am "$(COMMIT_MSG)"

rebase:
	rm -rf .git
	find . -type f -iname "*.go*" -exec sed -i '' -e "s%github.com/dimaskiddo/go-cron%$(REBASE_URL)%g" {} \;
	git init
	git remote add origin https://$(REBASE_URL).git

push:
	git push origin master

pull:
	git pull origin master
