# TODOs

- [X] Lets start by using drivers
- [X] Read `heph.fn.call()` and so on
- [ ] Hover `heph.fn.call()` and so on
- [X] Heph: driver should be inserted at load time
- [ ] Load hephconfig to really see stuff
- [ ] All uint conversions should be validated before, use some common function
- [X] Scope Variables
- [X] Scope Variables in function call
- [ ] Use chain in capabilites
- [ ] Accept struct and see how to work with symbols there
- [ ] Read the yaml to know it it will be always BUILD
- [ ] Complete copmpletion based on new hierarchy and symopls
- [ ] Read all files of current workdir



## Future

1. Change how we parse stuff. We should probably parse scope by scope, so we can have all data first:
  - Variables -> fun def -> fun call
  - This way we recursively go from scope from scope knowing current (and outer) scope variables and function definitions
2. Instead of fallback to first BUILD in directory, we can search for that function definition
