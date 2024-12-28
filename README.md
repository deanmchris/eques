Overview
--------

Eques is an open-source UCI compatible chess engine.

The name comes from the Latin word for knight.

Ratings
-------

When discussing an engine's (or human chess player's) strength, it's important to remember that the Elo is always relative to one's testing conditions. One tester may estimate an engine's strength to be 2300 for example, while another may get 2400. Neither tester is "wrong" per se, but they both likely have a different pool of opponets, different hardware, different time controls, etc.

With that said, an estimate of each versions rating is included below.

| Version     | Estimated Rating (Elo) |
| ----------- | -----------------------|
| 1.0.0       | 1700                   |


Installation
------------

Builds for Windows, Linux, and MacOS are included with each release of Eques. However, if you
prefer to build Eques from scratch, the steps to do so are outlined below:

- Visit the [Golang download page](https://golang.org/dl/), and install Golang using the download
package appropriate for your machine. Make sure to add the `go\bin` file to your path.

- Navigate to `eques/eques` and run `go build` to compile a binary for your machine. Go offers several
  environment variables to target specfic operating systems, architectures, and instruction sets.
  I can't find a good unified source listing the available values, but see the makefile in the project
  for an example of some supported options.

Alternatively, a makefile is included with this project to make compilation easier. The targets provided
are outlined below.

- run `make build`/`make build-windows` to build four different builds: one that works
on all AMD 64 architectures (default), and three that work with popcnt, avx2, and avx512 respectively. These
are not the only extended instruction sets supported by each respective build, as the 
Go compiler offers the ability to compile to diffent levels, rather than specfic 
microarchitectures. See [here](https://github.com/golang/go/wiki/MinimumRequirements#amd64) for more details.

- run `make build-all-default`/`make build-all-default-windows` to build default AMD 64 builds for macOS, linux, and windows.

- run `make build-all`/`make build-all-windows` to build default AMD 64 builds for macOS, linux, and windows, as well as popcnt, avx2, and avx512 builds for each OS.

- run `make clean-build`/`make clean-build-windows`, `make clean-all-default`/`make clean-all-default-windows`, `make clean-all`/`make clean-all-windows` to clean up the output from
the above commands, respectively.

Usage
-----

Like many chess engines, Eques does not provide it's own chess GUI, but supports something
known as the [UCI protocol](http://wbec-ridderkerk.nl/html/UCIProtocol.html). This protocol allows chess engines, like Eques, 
to communicate with different chess GUI programs.

So to use Eques, it's reccomend you install a dedicated chess GUI. Popular free ones include:

* [Arena](http://www.playwitharena.de/)
* [Scid](http://scidvspc.sourceforge.net/)
* [Cute-chess](https://cutechess.com/) 
* [Banksia](https://banksiagui.com/)
* [En Croissant](https://encroissant.org/)

Once you have a program downloaded, you'll need to follow that specfic programs guide on how to install a chess engine. When prompted 
for a command or executable, direct the GUI to the Golang exectuable you built, or one of the executables included with the specfic
release of Eques you're interested in.

Current Features
--------

* Engine
    - [Bitboard representation](https://www.chessprogramming.org/Bitboards)
    - [Magic bitboards for slider move generation](https://www.chessprogramming.org/Magic_Bitboards)
    - [Zobrist hashing](https://www.chessprogramming.org/Zobrist_Hashing)
* Search
    - [Negamax search framework](https://www.chessprogramming.org/Negamax)
    - [Alpha-Beta pruning](https://en.wikipedia.org/wiki/Alpha%E2%80%93beta_pruning)
    - [MVV-LVA move ordering](https://www.chessprogramming.org/MVV-LVA)
    - [PV move ordering](https://www.chessprogramming.org/Principal_Variation)
    - [Check extensions](https://www.chessprogramming.org/Check_Extensions)
    - [Quiescence search](https://www.chessprogramming.org/Quiescence_Search)
    - [Time-control logic supporting classical, rapid, bullet, and ultra-bullet time formats](https://www.chessprogramming.org/Time_Management).
    - [Repetition detection](https://www.chessprogramming.org/Repetitions)
* Evaluation
    - [Material evaluation](https://www.chessprogramming.org/Material)
    - [Tuned piece-square tables](https://www.chessprogramming.org/Piece-Square_Tables)
    - AdaGrad gradient descent [Texel Tuner](https://www.chessprogramming.org/Texel%27s_Tuning_Method)

See `docs/testing.md` for a log of the specfic features I've implemented, as well as their recorded Elo gains from testing. 
    
Changelog
---------
 
The changelog of features can be found in `docs/changelog.md`.
 
License
-------
 
Eques is licensed under the [MIT license](https://opensource.org/licenses/MIT).

Help & Support
--------------

I'm always happy and open to hearing any bug reports, typo corrections, code-cleanup or any other suggesstions.

Misc
----

If you enjoy my work with Eques, be sure to check out his cousin, [Blunder](https://github.com/deanmchris/blunder/tree/main)!