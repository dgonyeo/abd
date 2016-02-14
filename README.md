# appc Binary Discovery (abd)

Working copy of the spec: [abd.md](abd.md)

Historical design doc is here: https://docs.google.com/document/d/17G4FOroAW0utSIvuHqwM9SIPdkAyJJYqNdgBdrIztu0/edit#

## Usage

```
./build
./bin/abd --config-dir defaultConfigs --strategy-dir bin/strategies discover example.com/package,label1=foo,label2=bar
```
