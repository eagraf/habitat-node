# Test Suite

This test suite allows for different implementations of `entities.TransitionSubscriber` to receive
sequences of transitions.

## Running

Add Make rules for each new test you add. For example, the process manager test is run with:

```make -C test-suite test-process-manager```

To run the test-suite binary, do:

```<path-to-bin-location>/test-suite <subscriber_type> <test_file>```