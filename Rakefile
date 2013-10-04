ENV["GOPATH"] = ENV["PWD"]

task :test do
  sh "go test justworks/translog"
end

task :bench do
  sh "go test -test.bench 'KeyValueExtraction'  justworks/translog"
end

task :install do
  sh "go install justworks/translog/translog"
end

task :prep do
  sh "go get github.com/stretchr/testify/assert"
end

task :fmt do
  sh "gofmt -tabs=false -tabwidth=2 -w=true -l=true src/justworks/translog"
end

task :default => [:test, :install]
