# Koi

Koi is a language that seeks to make many traditional bugs impossible by preventing them at the language level.
Each of these are discussed in more detail below.

1. **Prevent all runtime errors.** Runtime errors are, by definition, unexpected and either have to be caught and emit and error under a sitution that can't be handled safely or cause the program to blow up.
2. **No garbage collector, but no manual memory management either.** It's unsafe to manage memory manully, but we also don't want the overhead of a garbage collector.
3. **Interoperability with C.** This is critical to making sure we can use existing libraries and also makes the lanaguage a lot easier if we can offload the lowest level logic to C.
4. **All objects are also interfaces.** Any object can be provided for a type if it fits the receiving interface.
5. **First class testing.** Language constructs for dealing with tests, assertions, mocks, etc.
6. Sum types handle behaviors like errors.

## Avoding Runtime Errors

There is a differnce between detecting and avoiding. We have to be careful avoiding doesn't become a burdon.

1. Nil derefernecing.
2. Divide by zero.
3. Overflow and underflow.
4. NaN and infinities.
5. All matching must be exhaustive?
6. Array and map out of bounds.
7. Casting to an invalid type.
8. Signals and other interupts.
9. Run out of of stack depth?

## Avoiding Logic Errors

1. Explicit order of operations.
2. Zero out memory.
3. Explicit mutability.
4. No jumps/gotos (including breaking).
5. No operator overloading.
6. No type overloading.
7. Infinite loops.
8. Infinite recursion?

## Processes

Memory cannot be shared between processes.
Launching a process returns a different type (ie. Process[MyObject]) that itself provides the API for syncronizing calls.
Any value that attempts to cross a process boundary must implement a `Copy` interface.

## Memory Management

Reference counting.

## Types and Domains

## All Objects Are Interfaces

## Testing

- Tests
- Assertions
- Mocks

## Language Constructs

### Data Types

### Variables

### Functions

### Sum Types

```
type NumberOrString = number | string

type MultiValues = (number, bool) | None

type NamedTypes = @high (number, number) | @low (number, number) | number
```

```
func static[doStuff] :good number | :bad number | error {

}

func {
  const result = match static[doStuff] {
    :good n number {
      n + 5
    }
    :bad number {
      -1
    }
    error {
      0
    }
  }
}
```

### Objects

### Control Flow

### Errors

Erorrs are just a return type. Auto snapshotting?

## Package Management
