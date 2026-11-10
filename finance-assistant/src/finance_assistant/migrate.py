from __future__ import annotations

import asyncio
from importlib.resources import files

import psycopg

from .config import get_settings


async def migrate() -> None:
    settings = get_settings()
    sql = files("finance_assistant").joinpath("sql/0001_initial.sql").read_text(encoding="utf-8")
    async with await psycopg.AsyncConnection.connect(settings.database_url, autocommit=True) as connection:
        await connection.execute(sql)


if __name__ == "__main__":
    asyncio.run(migrate())
