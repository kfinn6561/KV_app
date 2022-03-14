//Package KVStore includes all the dealings with the Key value store
//Simple functions are presented to the user, but internally it uses an actor model to ensure confinement of the data.
//For this reason the store must be initialised with the Startup() method, which starts the store actor.
//any functions that interact directly with the memory are prefaced by the word "direct" and should only be accessed via the actor.
//These are not exported from the package so it should be safe.
package KVStore
