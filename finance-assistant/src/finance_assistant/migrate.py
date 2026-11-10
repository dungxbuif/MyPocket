from __future__ import annotations

import asyncio
from importlib.resources import files

import psycopg
from psycopg import sql

from .config import get_settings


async def migrate() -> None:
    settings = get_settings()
    migration_sql = (
        files("finance_assistant")
        .joinpath("sql/0001_initial.sql")
        .read_text(encoding="utf-8")
    )
    async with await psycopg.AsyncConnection.connect(
        settings.database_url, autocommit=True
    ) as connection:
        await connection.execute(
            sql.SQL("CREATE SCHEMA IF NOT EXISTS {}").format(sql.Identifier(settings.database_schema))
        )
        await connection.execute(
            sql.SQL("SET search_path TO {}").format(sql.Identifier(settings.database_schema))
        )
        await connection.execute(migration_sql)
        result = await connection.execute(
            "SELECT to_regclass(%s)",
            (f"{settings.database_schema}.assistant_run_jobs",),
        )
        if (await result.fetchone())[0] is None:
            raise RuntimeError("assistant schema verification failed")


if __name__ == "__main__":
    asyncio.run(migrate())
