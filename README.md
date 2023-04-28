# ent-graphql-example

The code for https://entgo.io/docs/tutorial-setup

## Installation

```console
git clone git@github.com:a8m/ent-graphql-example.git
cd ent-graphql-example
go run ./cmd/todo/
```

## How to reproduce the bug

In GraphQL the following works okay:
```
query noIssueWithNullAfter {
  node(id: "4294967297") {
    ... on User {
      id
      name
      subordinates(first: 2, after: null) {
        edges {
          cursor
          node {
            id
            name
            todos(first: 4) {
              edges {
                cursor
                node {
                  id
                  text
                }
              }
            }
          }
        }
      }
    }
  }
}
```

But the following does not correctly fetch the todos:
```
# The todo edges that get returned are empty, even though the same User
# in the first query shows nodes.
query issueWithAfterFirst {
  node(id: "4294967297") {
    ... on User {
      id
      name
      subordinates(first: 2, after: "gaFpzwAAAAEAAAAC") {
        edges {
          cursor
          node {
            id
            name
            todos(first: 4) {
              edges {
                cursor
                node {
                  id
                  text
                }
              }
            }
          }
        }
      }
    }
  }
}
```

## Generated SQL
```
SELECT `users`.`id`, `users`.`name`
FROM `users`
WHERE `users`.`id` = 4294967297
LIMIT 2;
```
```
WITH `src_query` AS (SELECT `users`.`id`, `users`.`name`, `users`.`manager_id`
FROM `users`
WHERE `users`.`id` > 4294967298
AND `users`.`manager_id` IN (4294967297)),
`limited_query` AS (SELECT *,
(ROW_NUMBER() OVER (PARTITION BY `manager_id` ORDER BY `id` ASC)) AS `row_number`
FROM `src_query`)
SELECT `id`, `name`, `manager_id`
FROM `limited_query` AS `users`
WHERE `users`.`row_number` <= 3;
```
```
WITH `src_query` AS (SELECT `todos`.`id`, `todos`.`text`, `todos`.`owner_id`, `todos`.`todo_parent`
FROM `todos`
WHERE `todos`.`id` > 4294967298 -- 👈 this is a user ID not a todo ID! 🚨
AND `todos`.`owner_id` IN (4294967299, 4294967300, 4294967301)),
`limited_query` AS (SELECT *,
(ROW_NUMBER() OVER (PARTITION BY `owner_id` ORDER BY `id` ASC)) AS `row_number`
FROM `src_query`)
SELECT `id`, `text`, `owner_id`, `todo_parent`
FROM `limited_query` AS `todos`
WHERE `todos`.`row_number` <= 3;
```

Then, open [localhost:8081](http://localhost:8081) to see the GraphQL playground. 

## Report Issues / Proposals

The issue tracker for this repository is located at [ent/ent](https://github.com/ent/ent/issues).

## Join the ent Community
In order to contribute to `ent`, see the [CONTRIBUTING](https://github.com/ent/ent/blob/master/CONTRIBUTING.md) file for how to go get started.  
If your company or your product is using `ent`, please let us know by adding yourself to the [ent users page](https://github.com/ent/ent/wiki/ent-users).

## License
This project is licensed under Apache 2.0 as found in the [LICENSE file](LICENSE).
