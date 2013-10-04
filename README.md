# General

Lightweight log/event processor.

Goals:

  * fast
  * easy to deploy, zero dependencies
  * easy configuration

## Getting started

### Prebuild binaries (TODO)

<pre>
[shell1] $ wget TODO
[shell1] $ echo "[input.FileReaderPlugin]
filename=file0

[filter.KeyValueExtractor]

[output.StdoutWriterPlugin]" > translog.conf

[shell1] $ chmod a+rx translog
[shell1] $ ./translog -config demo.conf

[sheel2] $ echo "foo=bar baz baz baz" >> file0

[shell1]
2013/09/27 18:01:54 [inputProcessor] register *translog.FileReaderPlugin
2013/09/27 18:01:54 [filterChain] register *translog.KeyValueExtractor
2013/09/27 18:01:54 [outputProcessor] register *translog.StdoutWriterPlugin
2013/09/27 18:01:54 [EventPipeline] 0 msg, 0 msg/sec, 0 total, 5 goroutines
2013/09/27 18:02:04 [EventPipeline] 0 msg, 0 msg/sec, 0 total, 6 goroutines
<#Event>  Host: bender.local
  Source: file://file0
  InitTime: 2013-09-27 18:02:10.40223812 +0200 CEST
  Time: 2013-09-27 18:02:10.402238172 +0200 CEST
  Raw:
foo=bar baz baz baz
  Fields:
     [foo] bar
2013/09/27 18:02:14 [EventPipeline] 1 msg, 0 msg/sec, 1 total, 6 goroutines
CTRL-C
</pre>

### Compile

<pre>
$ git clone https://github.com/invadersmustdie/translog.git
$ cd translog
$ rake prep
$ rake
$ ./bin/translog -h
</pre>

# Sample Configuration

<pre>
[input.FileReaderPlugin]
filename=/var/log/nginx/access.log

[filter.KeyValueExtractor]

[output.StatsdPlugin]
host=metrics.local
port=8125
proto=udp
field.0.raw=test.nginx.runtime:%{runtime}|ms
field.1.raw=test.nginx.response.%{RC}:1|c
</pre>

# FAQ

  Q: Why not using logstash?
  A: First of all, logstash is awesome. It's the tool of choice if you want to do enhanced log/event processing. translog does not intend to replace logstash at all. It just provides a lite alternative if you have want to have real simple log processing.

  Q: How about more plugins for translog?
  A: As translog tries to keep the number of dependencies down (to zero) I'm really careful about adding new plugins. If you need to work redis, elasticsearch, ... use logstash.

  Q: What's throughput can it handle?
  A: There is no real answer to this question because it depends on the filter and configuration given. In general you will see a drop in througput when using complex regex patterns.

  On my local box (MBP2012 2.7GHz 8GB SSD) (-cpus=1) achieved the following numbers:

   * 1x FileReaderPlugin

<pre>
[input.FileReaderPlugin]
filename=tmp/source.file0
</pre>

-> 52540 msg/sec

   * 1x FileReaderPlugin + 1x KeyValueFilter

<pre>
[input.FileReaderPlugin]
filename=tmp/source.file0

[filter.KeyValueExtractor]
</pre>

-> 5168 msg/sec

   * 1x FileReaderPlugin + 1x KeyValueFilter + 1x FieldExtractor

<pre>
[input.FileReaderPlugin]
filename=tmp/source.file0

[filter.KeyValueExtractor]

[filter.FieldExtractor]
field.visitorId=visitorId=([a-zA-Z0-9]+);
field.varnish_director=backend="([A-Z]+)_
field.varnish_cluster=backend="[A-Z]+_([A-Z]+)_
</pre>

-> 4926 msg/sec

## Configuration & Plugins

### Input

#### TcpReaderPlugin

**Example**

<pre>
[input.TcpReaderPlugin]
port=8888
</pre>

#### FileReaderPlugin

**Example**

<pre>
[input.FileReaderPlugin]
filename=tmp/source.file0
</pre>

#### NamedPipeReaderPlugin

**Example**

<pre>
[input.NamedPipeReaderPlugin]
filename=tmp/source.pipe0
</pre>

### Filter

#### KeyValueFilter

**Example**

<pre>
[filter.KeyValueExtractor]
# no configuration
</pre>

#### FieldExtractor

Extracts value from raw message and adds it event. Regex pattern named in first group (parenthesis) will be used.

**Example**

<pre>
[filter.FieldExtractor]
debug=true
field.visitorId=visitorId=([a-zA-Z0-9]+);
</pre>

#### DropEventFilter

**Example**

<pre>
[filter.DropEventFilter]
debug=true
field.direction=c
msg.match=^foo
</pre>

#### ModifyEventFilter

**Example**

<pre>
[filter.ModifyEventFilter]
debug=true
field.remove.list=cookie,set-cookie
field.remove.match=^cache
</pre>

### Output

#### StdoutWriterPlugin

**Example**

<pre>
[output.StdoutWriterPlugin]
# no configration
</pre>

#### GelfWriterPlugin

**Example**

<pre>
[output.GelfWriterPlugin]
debug=true
host=localhost
port=12312
proto=tcp
</pre>

#### NamedPipeWriterPlugin

#### StatsdPlugin

**Example**

<pre>
[output.StatsdPlugin]
debug=true
host=foo.local
port=8125
proto=udp
field.0.raw=test.nginx.runtime:%{runtime}|s
field.1.raw=test.nginx.response.%{RC}:1|c
field.2.raw=test.nginx.%{foo}.%{bar}:1|c
</pre>

#### NetworkSocketWriter

**Example**

<pre>
[output.NetworkSocketWriter]
debug=true
host=localhost
port=9001
proto=tcp
</pre>
