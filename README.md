## midi-daemon

> a daemon that automatically records and saves midi in the background. useful if you spontaneously improvise often. 

### usage

1. run `./midi-recorder-daemon -show-ports` while your midi device is connected, it should show a list of entries in a format like `<number> | <device name>`, you should then note down the `<number>` or a unique part of `<device name>` so you can reference it later:
2. run either:
    1. `./midi-recorder-daemon -port-name <unique part of device name>` to search for midi devices whose names contain the `<unique part>` specified.
    1. `./midi-recorder-daemon -port-number <port number>` to search for a midi device with the exact port number specified.
3. if you want to specify the output location (by default it is `./output`), add `-output <path>` to the command.
4. leave the command running and the daemon will automatically start recording when midi messages are received.

here are the full command flags of `midi-recorder-daemon`:

```
Usage of ./midi-recorder-daemon:
  -instrument string
        the instrument to use in the midi file. (default "Piano")
  -meter-denominator int
        the denominator of the time signature. (default 4)
  -meter-numerator int
        the numerator of the time signature. (default 4)
  -output string
        the path to the folder where all midi recordings will be saved. (default "output")
  -port-name string
        search for a midi port by a keyword in its lowercased name, this flag is mutually exclusive with '-port-number'.
  -port-number int
        search for a midi port by its port number, this flag is mutually exclusive with '-port-name'. (default -1)
  -port-poll-timeout int
        seconds to wait between polling if a midi port exists. (default 5)
  -show-ports
        show midi ports available without recording, if you specify this flag, you will not need to specify any other flags.
  -tempo float
        the tempo of the midi file. (default 120)
  -timeout int
        how many seconds to wait before saving the current midi recording. (default 10)
```

