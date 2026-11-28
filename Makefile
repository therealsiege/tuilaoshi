.PHONY: build run test fmt check lint golden recordings launcher demonart freqlist proofsheet sweep reminder reminder-off ios ios-core ios-test ios-archive clean

GOLANGCI_LINT = go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0

build:
	go build -o cmd/mandarin/mandarin ./cmd/mandarin

# Run from the repo root so the app finds ./wiki.
run: build
	./cmd/mandarin/mandarin

test:
	go test ./...

fmt:
	gofmt -w cmd internal

# check runs everything CI runs.
check:
	@unformatted=$$(gofmt -l cmd internal); \
	if [ -n "$$unformatted" ]; then \
		echo "These files are not gofmt-formatted. Run 'make fmt':"; echo "$$unformatted"; exit 1; \
	fi
	./scripts/check-file-length.sh
	go vet ./...
	go build ./...
	go test ./...

lint:
	$(GOLANGCI_LINT) run ./...

# Rewrite screen snapshots (testdata/*.golden) after an intended UI change; review the diff.
golden:
	UPDATE_GOLDEN=1 go test ./internal/ui/...

# Record docs/recordings/*.gif from docs/tapes (needs vhs: brew install vhs).
recordings:
	@for t in docs/tapes/*.tape; do vhs "$$t" || exit 1; done

# Put TuiLaoShi.app on the Desktop (double-click to start the tutor).
launcher:
	./scripts/install-launcher.sh

# Paint demon portraits with Gemini Flash Image (GEMINI_API_KEY in ~/.config/tuilaoshi/env).
# Pass ARGS, e.g. make demonart ARGS="-only 白骨精 -force".
# Import a frequency word list (mylingua xlsx) into the private wiki, e.g.
# make freqlist IN="$$HOME/Desktop/list.xlsx".
freqlist:
	go run ./cmd/freqlist -in "$(IN)" -wiki wiki

# The iOS app: the Go engine as an xcframework (needs gomobile), then its
# tests in the iPhone simulator (needs XcodeGen).
ios-core:
	PATH="$$HOME/go/bin:$$PATH" gomobile bind -target=ios,iossimulator -o ios/build/TuiLaoShiCore.xcframework ./mobile

# Build the app and open it in the iPhone 17 Pro simulator.
ios: ios-core
	cd ios && xcodegen generate && xcodebuild build -project TuiLaoShi.xcodeproj -scheme TuiLaoShi \
		-destination 'platform=iOS Simulator,name=iPhone 17 Pro' -derivedDataPath DerivedData -quiet
	xcrun simctl boot 'iPhone 17 Pro' 2>/dev/null || true
	xcrun simctl install 'iPhone 17 Pro' ios/DerivedData/Build/Products/Debug-iphonesimulator/TuiLaoShi.app
	xcrun simctl launch 'iPhone 17 Pro' com.therealsiege.tuilaoshi

ios-test: ios-core
	cd ios && xcodegen generate && xcodebuild test -project TuiLaoShi.xcodeproj -scheme TuiLaoShi \
		-destination 'platform=iOS Simulator,name=iPhone 17 Pro' -derivedDataPath DerivedData -quiet

# A Release archive for App Store Connect (ios/build/TuiLaoShi.xcarchive),
# then the .ipa from it (ios/build/export) per ios/ExportOptions.plist. The
# build number is the commit count. Signing needs the team's account in
# Xcode; UNSIGNED=1 checks the archive builds without it and skips the export.
IOS_BUILD_NUMBER = $(shell git rev-list --count HEAD)
ios-archive: ios-core
	cd ios && xcodegen generate && xcodebuild archive -project TuiLaoShi.xcodeproj -scheme TuiLaoShi \
		-configuration Release -destination 'generic/platform=iOS' -archivePath build/TuiLaoShi.xcarchive \
		-derivedDataPath DerivedData -quiet CURRENT_PROJECT_VERSION=$(IOS_BUILD_NUMBER) \
		$(if $(UNSIGNED),CODE_SIGNING_ALLOWED=NO,-allowProvisioningUpdates)
	$(if $(UNSIGNED),@echo "Unsigned archive: skipping the export.",cd ios && xcodebuild -exportArchive \
		-archivePath build/TuiLaoShi.xcarchive -exportOptionsPlist ExportOptions.plist \
		-exportPath build/export -allowProvisioningUpdates)

# A daily study reminder at AT (a macOS notification), e.g. make reminder AT=19:00.
reminder:
	./scripts/install-reminder.sh "$(or $(AT),19:00)"

reminder-off:
	./scripts/install-reminder.sh off

# Open every screen of the real game at 80x24 and 120x45 (tmux) and fail on
# a missing title or a cut line. Takes a few minutes.
sweep:
	./scripts/sweep.sh

# Every Chinese line of the content in one sheet for a native speaker, e.g.
# make proofsheet OUT="$$HOME/Desktop/proofsheet.md".
proofsheet:
	go run ./cmd/proofsheet -out "$(OUT)"

demonart:
	zsh -c 'source ~/.config/tuilaoshi/env 2>/dev/null; go run ./cmd/demonart $(ARGS)'

clean:
	rm -f cmd/mandarin/mandarin
