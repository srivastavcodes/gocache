# gocache

gocache is a small caching library that provides a thread-safe lru implementation built on top of a custom doubly linked list.

The library also contains a simplelru implementation which serves as a lightweight option with only core methods and no
synchronization.

## Features

- All the standard caching operations:
    - Add, Remove, ContainsOrAdd, PeekOrAdd, GetOldest, RemoveOldest, etc.

## Status

I am going to add a few more caching strategies as I learn about them and for now the lru is ready for usage.

Feel free to create a pull request if you want to add a new feature or fix a bug.
