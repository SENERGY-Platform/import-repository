# Criteria aspect ids, and the parameter that ANDs them

## Scope

Holds for the criteria filter of `GET /import-types` and for the derived `criteria`
array the filter runs on, from the change that introduced
`and_combine_criteria_aspect_ids`. Anything that touches `lib/model`'s aspect
normalization, the `criteria` documents in mongo, or the aspect resolution in the
controller is in scope.

Three neighbouring cases look the same and are not, and none of them can be
assumed from this document:

- **The aspect list of a content variable is not the aspect list of a criterion.**
  Both are called `aspect_ids` and both hold aspect urns. A content variable
  *enumerates what it carries*; a criterion *demands that a variable carry them*.
- **ANDing inside one criterion is not ANDing the criteria list.** The criteria
  list has always been an AND at import-type level: `[{aspects:[a]},
  {aspects:[b]}]` matches a type whose one variable carries `a` and whose other
  carries `b`. `[{aspects:[a,b]}]` with the parameter set asks for a *single*
  variable carrying both, and the same import type does not match.
- **The default is unchanged.** Without the parameter the list is an OR and no
  aspect subtree is resolved. A caller that relied on expanding the subtree itself
  keeps working, and must keep doing it.

## The criteria index

`SetImportType` stores an import type as `ImportTypeWithCriteria`: the import type
plus a flattened `criteria` array holding one entry per content variable of the
whole `output` tree, each `{function_id, aspect_ids}`. A variable with no function
and no aspect still gets an entry, so an import type always has at least one.

The array is derived, never written by a client, and it is replaced wholesale on
every write — `SetImportType` uses `ReplaceOne`, so a field that a previous version
wrote and this one does not disappears with the next write of that import type.

## The filter

One `$elemMatch` per requested criterion, joined by `$and`, so each criterion has
to be satisfied by *one* content variable while different criteria may pick
different variables:

```js
{"$and": [
  {"criteria": {"$elemMatch": { /* criterion 1 */ }}},
  {"criteria": {"$elemMatch": { /* criterion 2 */ }}}
]}
```

The aspect part of one criterion is a list of id **sets**: a variable matches a set
by carrying any of its ids, and it has to match every set. Both modes reduce to
that one shape, which is why the database does not know about the parameter:

- OR (default) — one set, holding the ids as given.
- AND — one set per named aspect, holding that aspect and its descendants.

The sets are ANDed *inside* the `$elemMatch`, which needs an explicit `$and`
because the same field key cannot appear twice in one document. Splitting them into
two `$elemMatch` predicates instead would let two different content variables
satisfy them, which is a different and wider query. `addAspectIdSetsToCriteriaFilter`
keeps the zero, one and many cases apart; the single case is written as a plain
field predicate so the common query stays readable.

## Where the aspect subtree is resolved, and why not in the database layer

The controller resolves it, in `importTypeQueryOptions`, before the database sees
the query. `lib/database` performs no network calls, and the aspect hierarchy is
not this service's data — it has to be fetched from the device-repository. That is
also why the two option structs exist:

- `model.ImportTypeListOptions` is the request as it arrives, with
  `Criteria []ImportTypeFilterCriteria` and the `AndCombineCriteriaAspectIds` flag.
- `model.ImportTypeQueryOptions` is what the database evaluates, with
  `Criteria []ImportTypeCriteriaQuery` carrying the resolved `AspectIdSets`.

The alternative — one struct, the flag passed down and an aspect provider injected
into `lib/database/mongo` — was rejected: it puts an outbound HTTP dependency into
the storage layer. The device-repository can do exactly that in its own mongo
layer, because its aspect nodes sit in the same database; here they do not, and
copying the shape without the precondition is what makes the layering wrong.

`getAspectNode` caches per node id for one minute (`aspect-nodes.<id>`,
`service-commons/pkg/cache`), so a criteria list that names the same aspect
repeatedly costs one call. An aspect the device-repository answers with 404 is
**tolerated**, logged at WARN, and treated as a node without descendants — it then
covers only itself, so a typo narrows the answer instead of failing the request.
This mirrors what the device-repository does with an unknown aspect in its own
filter.

## The deprecated aspect_id and the migration

`ContentVariable.AspectId` is deprecated in favour of `AspectIds`. The
normalization lives in `lib/model/content_variable_aspects.go`, not in the
controller, because two callers need it and only one of them is a controller:

- On write, `SetContentVariableAspectIdsOnWrite` folds a non-empty `AspectId` into
  `AspectIds`. Everything behind that point — validation, the criteria array,
  `ExtendImportType` — reads `AspectIds` only.
- On read, `SetContentVariableAspectIdsOnRead` folds the same way first (documents
  written before the change carry `AspectId` alone) and then sets `AspectId` to the
  alphabetically first entry, so a client that knows only the old field keeps
  working.
- A filter-criterion carries the same deprecated field, because the shared
  `ImportTypeFilterCriteria` declares both. `FilterCriteriaAspectIds` folds it into
  the list in `importTypeQueryOptions`, so a criterion that names only `aspect_id`
  filters instead of being dropped. It used to be an unknown json field and was
  silently ignored.
- `migrateImportTypeCriteria` in `lib/database/mongo/import_types.go` calls the
  write normalization directly. It re-saves stored import types whose documents
  lack `criteria` **or** whose criteria entries lack `aspect_ids`, and it runs in
  `mongo.New`, which cannot reach the controller.

Two consequences of that migration. A stored criteria entry that carried the old
single `aspect_id` no longer does after it runs, so **rolling back to a binary that
filters on `aspect_id` returns empty results rather than failing** — the field it
looks for is gone from every migrated document. And an import type whose content
variables have no aspects at all is migrated too, because its entries have no
`aspect_ids` field either; that is harmless and idempotent, but it means the
migration touches every document once rather than only the ones with aspects.

## Where the types live

`lib/model` declares the import-type shapes as aliases of the shared model in
`github.com/SENERGY-Platform/models/go/models`, so there is one definition per
shape and the names this service's client exposes stay put:

| `lib/model`                | shared model                    |
| -------------------------- | ------------------------------- |
| `ImportType`               | `models.ImportType`             |
| `ContentVariable`          | `models.ImportContentVariable`  |
| `ImportConfig`             | `models.ImportTypeConfig`       |
| `ImportTypeFilterCriteria` | `models.ImportTypeFilterCriteria` |
| `Type` and its constants   | `models.Type`                   |

What stays local is what the shared model has no opinion about: the derived
`ImportTypeExtended`, the two option structs, `ImportTypeCriteriaQuery`, and the
aspect normalization above. The shared model calls the config declaration
`ImportTypeConfig` and reserves `ImportConfig` for the config *values* of an import
instance, which this service never handles — the alias keeps the local name.

## Reading the parameter

`and_combine_criteria_aspect_ids` is parsed with `strconv.ParseBool`, absent means
`false`, and an unparsable value is a 400 rather than a silent default — the same
handling as `limit` and `offset` in the same handler. The Go client sends it only
when true, so a generated request stays identical to what it was before this
change.
