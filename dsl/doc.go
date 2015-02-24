/*
Cicerone - a lager connoisseur

lager emits logs
chug pretty prints logs in real time
cicerone sifts through those logs and analyzes them

The dsl package has a number of different nouns:

- Entry: Cicerone's representation of a log line
- Entries: an ordered list of entries, can be filtered and grouped
- GroupedEntries: a collection of Entries grouped by an arbitrary key
- EntryPair: a set of two entries typically used to represent a period of time between two events of interest
- EntryPairs: a collection of EntryPair.  From this one can generate:
    - DTStats: a set of statistics describing a collection of EntryPairs
    - Durations: an array of time.Durations with some convenience helpers
- TimelineDescription: a collection of TimelinePoints used to construct a timeline
- Timeline: combines a TimelineDescription with an Entries -- represents the timeline associated with a particular object flowing through the logs
- Timelines: a pile of logs will have several timelines in them.  These are collected into a Timelines object.
- Matchers: matchers take an Entry and return a boolean
- Getters: getters take an Entry and pull data out of it

Using these nouns and their attendant verbs one can use Cicerone to quickly slice and dice a collection of log lines.  The resulting data can then be visualized with the viz package.
*/
package dsl
