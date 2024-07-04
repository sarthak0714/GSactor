# GSactor (Game server simulation using Actor Model)

This project implements a game server (some sort of PvE multiplayer BR) simulation using Go. It demonstrates the power of Go's concurrency model and showcases techniques for efficient game state management and parallel processing. There was an attempt to use actor model.

## Learning the Actor Model
I created this project to learn and apply the actor model, which simplifies managing concurrent processes by encapsulating state and behavior within actors. Each actor has a mailbox for receiving messages, and they interact through message passing. Actor does not have state so no locks, it is simple lightweight and fast.

## Key
1. **Actor Model**: Each game entity (Player, Enemy, Projectile) is an actor that can receive and process messages.
2. **Spatial Partitioning**: A grid system divides the game world into cells for efficient collision detection.
3. **Worker Pool**: A pool of goroutines processes game updates in parallel.
4. **Concurrent Spawning**: Enemies are spawned concurrently to simulate dynamic game world population.

## Some benchmarks
```bash
Starting simulation with 100 players...
Update 0/1000: 100 players, 1 enemies, 12 projectiles
Update 100/1000: 100 players, 101 enemies, 20 projectiles
Update 200/1000: 100 players, 155 enemies, 4 projectiles
Update 300/1000: 100 players, 155 enemies, 4 projectiles
Update 400/1000: 100 players, 155 enemies, 3 projectiles
Update 500/1000: 100 players, 154 enemies, 6 projectiles
Update 600/1000: 100 players, 152 enemies, 5 projectiles
Update 700/1000: 100 players, 152 enemies, 3 projectiles
Update 800/1000: 100 players, 152 enemies, 10 projectiles
Update 900/1000: 100 players, 151 enemies, 10 projectiles

Simulation completed in 16.586909268s
Final state: 100 players, 150 enemies, 16 projectiles
Updates: 1000
Operations per second: 6028.85
```