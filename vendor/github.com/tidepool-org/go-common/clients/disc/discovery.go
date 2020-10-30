/*
Package disc contains APIs and interfaces that are depended on as part of a discovery system
*/
package disc

type Discovery interface {
	Watch(service string) *Watch
	Publish(sl *ServiceListing)
}
