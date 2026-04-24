"""Database connection helpers.

All queries must be parametrized — never use f-strings in SQL.
"""

from typing import Any


def execute_query(
    cursor: Any,
    sql: str,
    params: tuple[Any, ...] = (),
) -> list[tuple[Any, ...]]:
    """Execute a parametrized SQL query and return all rows.

    Args:
        cursor: A psycopg2 cursor.
        sql: Parametrized SQL string using %s placeholders.
        params: Tuple of values to bind to the query placeholders.

    Returns:
        List of result rows as tuples.

    Example:
        rows = execute_query(
            cur, "SELECT * FROM invoices WHERE tenant_id = %s", (tenant_id,)
        )
    """
    cursor.execute(sql, params)
    return cursor.fetchall()  # type: ignore[no-any-return]
