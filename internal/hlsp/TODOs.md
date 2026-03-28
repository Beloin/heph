# TODOs

- [X] Lets start by using drivers
- [X] Read `heph.fn.call()` and so on
- [ ] Hover `heph.fn.call()` and so on
- [X] Heph: driver should be inserted at load time
- [ ] Load hephconfig to really see stuff
- [ ] All uint conversions should be validated before, use some common function
- [X] Scope Variables
- [ ] Scope Variables in function call
- [ ] Use chain in capabilites
- [ ] I do not like how we use scope rn. We work with symbols and not with a real scope tree
- [ ] Accept struct and see how to work with symbols there


## Future

1. Change how we parse stuff. We should probably parse scope by scope, so we can have all data first:
  - Variables -> fun def -> fun call
  - This way we recursively go from scope from scope knowing current (and outer) scope variables and function definitions
