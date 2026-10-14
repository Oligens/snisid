"""add NIF storage and indexed fuzzy identity search

Revision ID: 0013
Revises: 0012
Create Date: 2026-10-12
"""
from __future__ import annotations

from alembic import op
import sqlalchemy as sa

revision: str = "0013"
down_revision: str | None = "0012"
branch_labels: str | None = None
depends_on: str | None = None


def upgrade() -> None:
    op.execute("CREATE EXTENSION IF NOT EXISTS pg_trgm")
    op.add_column("citizens", sa.Column("nif", sa.String(length=32), nullable=True))
    op.add_column("citizens_read", sa.Column("nif", sa.String(length=32), nullable=True))
    op.create_index("ix_citizens_nif", "citizens", ["nif"])
    op.create_index("ix_citizens_read_nif", "citizens_read", ["nif"])
    op.execute(
        "CREATE INDEX ix_citizens_read_first_name_trgm "
        "ON citizens_read USING gin (lower(first_name) gin_trgm_ops)"
    )
    op.execute(
        "CREATE INDEX ix_citizens_read_last_name_trgm "
        "ON citizens_read USING gin (lower(last_name) gin_trgm_ops)"
    )
    op.execute(
        "CREATE INDEX ix_citizens_read_full_name_trgm "
        "ON citizens_read USING gin (lower(full_name) gin_trgm_ops)"
    )


def downgrade() -> None:
    op.execute("DROP INDEX IF EXISTS ix_citizens_read_full_name_trgm")
    op.execute("DROP INDEX IF EXISTS ix_citizens_read_last_name_trgm")
    op.execute("DROP INDEX IF EXISTS ix_citizens_read_first_name_trgm")
    op.drop_index("ix_citizens_read_nif", table_name="citizens_read")
    op.drop_index("ix_citizens_nif", table_name="citizens")
    op.drop_column("citizens_read", "nif")
    op.drop_column("citizens", "nif")
    # pg_trgm is intentionally retained because other services may share it.