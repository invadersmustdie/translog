ENV["GOPATH"] = ENV["PWD"]
VERSION_FILE = "src/justworks/translog/translog/version.go"

task :set_version do
  version = ENV["VERSION"]

  if !version
    # switch to default version
    version = File.read("VERSION").strip + "-" + `git rev-parse --short HEAD`.strip
  end

  new_content = ""

  File.readlines(VERSION_FILE).each do |line|
    if line =~ /^const APP_VERSION/
      new_content << %Q{const APP_VERSION = "#{version}"}
    else
      new_content << line
    end
  end

  File.open(VERSION_FILE, "w") do |fp|
    fp.puts new_content
  end
end

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

task :default => [:set_version, :test, :install]
