# Part 1 — Go Language Basics

← [Back to Chapter 1 index](README.md)

## Table of Contents

- [Why Go?](#why-go)
- [Setting Up](#setting-up)
- [Packages and `main`](#packages-and-main)
- [Variables, Types, and `:=`](#variables-types-and-)
- [For Loops — Go's Only Loop Keyword](#for-loops--gos-only-loop-keyword)
- [Functions and Multiple Return Values](#functions-and-multiple-return-values)
- [Function Literals (Anonymous Functions)](#function-literals-anonymous-functions)
- [Error Handling — No Exceptions](#error-handling--no-exceptions)
- [Defer](#defer)
- [Structs — Go's "Classes"](#structs--gos-classes)
- [Methods and Pointer Receivers](#methods-and-pointer-receivers)
- [Interfaces — Structural Typing](#interfaces--structural-typing)
- [Slices and Maps](#slices-and-maps)
- [Struct Tags](#struct-tags)

Next: [Part 2 — Concurrency](02-concurrency.md), where goroutines and channels are covered in depth.

---

## Why Go?

Go (also called **Golang**) is a compiled, statically-typed language created at Google. Compared to TypeScript and Python, the biggest mental shifts are:

| | TypeScript / Python | Go |
|---|---|---|
| **Execution** | Interpreted / JIT-compiled at runtime (Node.js, CPython) | Compiled ahead of time to a single native binary |
| **Typing** | TypeScript: static but erased at runtime. Python: dynamic. | Static and enforced at runtime — no `any`, no duck typing |
| **Errors** | `throw` / `try` / `catch` exceptions | Errors are ordinary return values you check explicitly |
| **Concurrency** | Single-threaded event loop (`async`/`await`, Promises) | Goroutines + channels, scheduled across real OS threads |
| **Object model** | Classes with inheritance | Structs + interfaces, no inheritance |
| **Package manager** | npm / pip / uv | Go modules (built into the toolchain) |
| **Null handling** | `null` / `undefined` / `None` | Every type has a "zero value"; pointers can be `nil` |
| **Dependencies to deploy** | `node_modules/`, a Node runtime, or a `.venv` + Python interpreter | None — `go build` produces one self-contained binary |

None of this makes Go "better" or "worse" — it's simply a different set of trade-offs, optimized for building simple, fast, and highly concurrent network services. Keep this table in mind as a reference as you read the rest of this chapter.

---

## Setting Up

Install Go (macOS via Homebrew):

```bash
brew install go
go version   # should print go1.21 or later
```

There is no separate package-manager binary to install (unlike `npm` for Node or `pip`/`uv` for Python) — dependency management is built into the `go` command itself, as you'll see in [Project Setup with Go Modules](04-api.md#project-setup-with-go-modules).

---

## Packages and `main`

Every Go file starts with a `package` declaration. A runnable program needs exactly one package called `main`, and that package needs a `func main()` — the entry point, conceptually the same role as a script's top-level code in Python or the file you point `node` at in Node.js.

```go
package main

import "fmt"

func main() {
	fmt.Println("Hello, Go!")
}
```

- `package main` — this file belongs to the `main` package (every `.go` file in a directory must declare the same package name).
- `import "fmt"` — `fmt` is a standard-library package for formatted I/O, roughly Go's equivalent of `console.log` (`fmt.Println`) plus Python's `str.format` (`fmt.Sprintf`).
- `func main()` — the entry point. Run it with `go run .`, or compile it to a binary with `go build`.

Unlike Node.js or Python, **unused imports and unused local variables are compile errors** in Go, not warnings — the compiler is intentionally strict about keeping code clean.

Most `.go` files in a real project **aren't** `package main` at all — they're shared library packages that other code imports by path, like `go-api/internal/items` or `go-api/internal/concurrency` in this tutorial's API. Only the small handful of files under `go-api/cmd/` are `package main` — see [Part 2](02-concurrency.md#cmd-and-internal-organizing-a-real-go-project) for why splitting code this way matters once a project has more than one runnable entry point.


## Variables, Types, and `:=`

Go is statically typed, but it has type inference, so you rarely have to write types out by hand — similar to TypeScript's `let x = 5` inferring `number`.

```go
var name string = "Go"   // explicit type
var age = 25             // inferred type (int)
count := 0                // shorthand declare + infer — only inside functions
count = count + 1
```

- `var` declares a variable, optionally with an explicit type. It works both inside and outside functions.
- `:=` is shorthand for "declare and initialize with an inferred type" — it can only be used inside a function body, not at package level. This is the form you'll see most often.
- There is no `let`/`const` split by mutability the way JS has `let` vs `const` — Go has a separate `const` keyword for compile-time constants, and everything else declared with `var`/`:=` is mutable.

### Zero values — Go's alternative to `null`/`undefined`

Every declared-but-unassigned variable in Go gets a **zero value** rather than being `null`, `undefined`, or `None`:

| Type | Zero value |
|---|---|
| `int`, `float64` | `0` |
| `string` | `""` (empty string) |
| `bool` | `false` |
| pointers, slices, maps, functions, interfaces | `nil` |

This is why you'll see `nil` used as Go's rough equivalent of `null`/`None`, but only for types that can actually be "nothing" (pointers, slices, maps, etc.) — a plain `int` or `string` can never be `nil`, it just defaults to `0` or `""`.

---

## For Loops — Go's Only Loop Keyword

Go has exactly one looping keyword, `for` — there's no separate `while`, `do...while`, or `foreach`. It covers four shapes:

```go
for i := 0; i < 3; i++ {   // classic three-clause form, like JS/TS/Python's `for i in range(...)`
	fmt.Println(i)
}

n := 0
for n < 3 {                // condition only — this is Go's `while`
	n++
}

for {                       // no condition at all — infinite loop, exited with `break`/`return`
	break
}

names := []string{"alice", "bob"}
for i, name := range names { // `range` iterates a slice, yielding (index, value)
	fmt.Println(i, name)
}
```

- `for range` also works on maps (yielding `key, value`), strings (yielding byte index, rune), and channels (yielding just the received value — see [Channels](02-concurrency.md#channels)).
- If you don't need one of the two values `range` yields, Go's convention is to discard it with the **blank identifier `_`** rather than naming a variable you'll never use — recall from [Packages and `main`](#packages-and-main) that an unused variable is a compile error, so `_` is how you explicitly say "I know this exists, I'm ignoring it":

  ```go
  for _, task := range tasks {
  	fmt.Println(task) // only the value is needed, so the index is discarded
  }
  ```

  You'll see this exact pattern in `go-api/internal/concurrency/fanout.go` and throughout this project's tests.

---

## Functions and Multiple Return Values

```go
func add(a int, b int) int {
	return a + b
}

func divide(a, b int) (int, error) {
	if b == 0 {
		return 0, fmt.Errorf("cannot divide by zero")
	}
	return a / b, nil
}

result, err := divide(10, 2)
```

- Parameter types come **after** the name (`a int`, not `int a`) — the reverse of TypeScript's `a: number`.
- Go functions can return **multiple values** — there's no need to bundle them into an object/tuple the way you might in TS (`{ result, error }`) or Python (`return result, error`). This is used everywhere in Go, most importantly for [error handling](#error-handling--no-exceptions).

---

## Function Literals (Anonymous Functions)

Functions are values in Go, just like in JS/TS or Python. A **function literal** is an unnamed `func` written inline — you can assign it to a variable, pass it as an argument, or call it immediately:

```go
add := func(a, b int) int { // a function literal assigned to a variable
	return a + b
}
fmt.Println(add(2, 3)) // 5

func(msg string) { // a function literal called immediately, right where it's defined
	fmt.Println(msg)
}("hello") // <- this trailing (...) is the call, with "hello" passed as msg
```

That last shape — define a function literal and call it in the same expression — is Go's equivalent of a JavaScript IIFE (`(function() { ... })()`). You'll see it combined with `go` in [The sync.WaitGroup](02-concurrency.md#the-syncwaitgroup):

```go
go func(t Task) {
	t.Run()
}(task)
```

Reading this left to right:

- `go` applies to the entire statement that follows it — it means "run this whole call as a new goroutine," not just "run `func`."
- `func(t Task) { t.Run() }` is the function literal itself — an anonymous function taking one `Task` parameter.
- `(task)` is the call — it immediately invokes that function literal, passing the current loop variable `task` in as the parameter `t`.

So the trailing `(task)` belongs to the function literal, not to `go`: the full expression is "call this anonymous function with `task`," and `go` just tells Go to run that call on its own goroutine instead of blocking.

---

## Error Handling — No Exceptions

This is the single biggest adjustment coming from TypeScript or Python. **Go has no `try`/`catch`/`throw`.** Instead, any function that can fail returns an `error` as its last return value, and the caller is expected to check it immediately:

```go
data, err := someFunction()
if err != nil {
	// handle the error — log it, return it, wrap it, etc.
	return err
}
// use `data` here — it's safe because err was nil
```

Compare this to what you're used to:

```typescript
// TypeScript
try {
  const data = someFunction();
  // use data
} catch (err) {
  // handle err
}
```

```python
# Python
try:
    data = some_function()
    # use data
except Exception as err:
    # handle err
```

The Go version has no hidden control flow — nothing "throws" and unwinds the stack automatically. Every error must be explicitly checked at the call site, which means Go code tends to have a lot of `if err != nil { ... }` blocks. It's more verbose, but it makes every possible failure point visible in the code, rather than hidden behind an implicit exception path.

> Go does have `panic`/`recover`, which behaves a bit like throwing/catching an exception, but it's reserved for truly unexpected, unrecoverable situations (e.g. a bug, an out-of-bounds array access) — not for ordinary error handling like "the request body was invalid JSON". You'll see `recover` used once in this tutorial's API, in the form of chi's `middleware.Recoverer` (see [How the API Works](04-api.md#how-the-api-works)), which exists purely as a safety net so a single bad request can't crash the whole server.

---

## Defer

`defer` schedules a function call to run right before the *enclosing function* returns — regardless of which `return` statement it hits, or whether it returns because of a `panic`. It's Go's answer to "always clean this up," without needing a `try`/`finally` block:

```go
func readFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close() // runs when readFile returns, no matter which path out is taken

	data, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
```

```typescript
// Roughly equivalent in TypeScript
function readFile(path: string): string {
  const file = openFile(path);
  try {
    return readAll(file);
  } finally {
    file.close(); // always runs, like defer
  }
}
```

A few things worth knowing:

- **Deferred calls run in LIFO order** (last deferred, first run) — if you `defer` three things, they run in reverse order of how you wrote them, similar to stacking multiple `finally` blocks.
- **Arguments are evaluated immediately**, only the call itself is delayed — `defer fmt.Println(x)` captures the current value of `x` right away, even though the print happens later.
- You'll see `defer` constantly in this tutorial's API code, e.g. `defer s.mu.Unlock()` in `go-api/internal/items/store.go` — lock the mutex, defer the unlock, and every `return` in the method automatically releases the lock without needing to repeat `mu.Unlock()` at every exit point.

---

## Structs — Go's "Classes"

Go has no classes and no inheritance. Instead, it has **structs**: plain data containers, similar to a TypeScript `interface`/`type` or a Python `dataclass`.

```go
type Item struct {
	ID          int
	Name        string
	Description string
}

item := Item{ID: 1, Name: "Sample", Description: "A sample item"}
fmt.Println(item.Name) // "Sample"
```

```typescript
// TypeScript equivalent
interface Item {
  id: number;
  name: string;
  description: string;
}
const item: Item = { id: 1, name: "Sample", description: "A sample item" };
```

```python
# Python equivalent (dataclass)
@dataclass
class Item:
    id: int
    name: str
    description: str
```

A key difference: struct fields starting with an **uppercase letter** are exported (public, visible outside the package), while lowercase fields are unexported (private to the package). There's no `public`/`private` keyword — visibility is determined entirely by the capitalization of the name. This is why you'll see `Item.Name` (uppercase, exported) but `Store.mu` (lowercase, package-private) in the API code.

---

## Methods and Pointer Receivers

Go doesn't have classes with methods baked in, but you can attach a method to any struct type using a **receiver**:

```go
type Counter struct {
	count int
}

// Pointer receiver — can modify the struct
func (c *Counter) Increment() {
	c.count++
}

// Value receiver — gets a copy, cannot modify the original
func (c Counter) Value() int {
	return c.count
}
```

- `(c *Counter)` is a **pointer receiver** — `c` is a pointer to the actual struct, so mutations inside the method affect the original. This is the closest equivalent to a regular mutating method (`this.count++`) in a TS/Python class.
- `(c Counter)` is a **value receiver** — `c` is a full copy of the struct, so changes inside the method are local and don't affect the caller's original value.
- The convention: if a method needs to modify the struct (or the struct is large and copying it would be wasteful), use a pointer receiver. Read-only methods on small structs can use value receivers.

You'll see this pattern throughout `go-api/internal/items/store.go` — e.g. `func (s *Store) Create(...)` uses a pointer receiver because it mutates the store.

---

## Interfaces — Structural Typing

Go interfaces describe a set of methods a type must implement — but unlike TypeScript's `implements` keyword, **Go interfaces are satisfied implicitly**. There's no explicit "this struct implements this interface" declaration; if a type has all the right methods, it automatically satisfies the interface.

```go
type Shape interface {
	Area() float64
}

type Circle struct {
	Radius float64
}

func (c Circle) Area() float64 {
	return 3.14159 * c.Radius * c.Radius
}

// Circle automatically satisfies Shape — no "implements Shape" needed
var s Shape = Circle{Radius: 2}
```

This is structurally similar to how TypeScript's structural typing works for object shapes (`interface Shape { area(): number }` — any object with an `area()` method matches), but Go extends that same idea to concrete types with methods, not just object literals.

The most important interface in this tutorial is `http.Handler` (from the standard library), which just requires a `ServeHTTP(w http.ResponseWriter, r *http.Request)` method — see [What is chi?](04-api.md#what-is-chi) for why that matters.

---

## Slices and Maps

- A **slice** (`[]T`) is Go's dynamic array — the rough equivalent of a TypeScript `Array<T>` / `T[]` or a Python `list`.
- A **map** (`map[K]V`) is Go's hash map — the equivalent of a TypeScript `Map<K, V>` / plain object, or a Python `dict`.

```go
names := []string{"alice", "bob"}       // slice literal
names = append(names, "carol")          // append (returns a new slice header)

ages := map[string]int{"alice": 30}     // map literal
ages["bob"] = 25                        // set
value, ok := ages["carol"]              // read + "does it exist?" check
```

The `value, ok := ages["carol"]` pattern is idiomatic Go: reading a missing map key doesn't throw or return `undefined`/`None` — it silently returns the value type's zero value, so `ok` is how you distinguish "key present with a zero value" from "key absent". You'll see this exact pattern in `go-api/internal/items/store.go`'s `Store.Get`.

---

## Struct Tags

You'll see snippets like this throughout the API code:

```go
type Item struct {
	ID   int    `json:"id"`
	Name string `json:"name" validate:"required,min=2,max=100"`
}
```

The backtick-delimited text after each field is a **struct tag** — a plain string attached to the field that libraries can read via reflection at runtime. Go itself does nothing with these strings; they're metadata that specific packages opt into reading:

- `encoding/json` (standard library) reads `json:"..."` tags to decide how to name each field when marshalling to/from JSON — the equivalent of how Pydantic maps Python field names to JSON keys, or how a TypeScript type just naturally matches the JSON shape.
- `go-playground/validator` (third-party) reads `validate:"..."` tags to know what rules to check — see [Validation with go-playground/validator](04-api.md#validation-with-go-playgroundvalidator).

This tag-based approach is how Go achieves declarative, schema-like behavior without a fluent builder API like Zod's `z.object({...})` or a base class like Pydantic's `BaseModel` — the "schema" is just annotations on a plain struct.

---

Next: [Part 2 — Concurrency](02-concurrency.md)
