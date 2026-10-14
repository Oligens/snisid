from __future__ import annotations

from datetime import date
from types import SimpleNamespace

import pytest
from pydantic import ValidationError
from sqlalchemy.dialects import postgresql

from services.identity.search import (
    IdentitySearchFilters,
    _aggregate_institutional_records,
    build_identity_search_statement,
)
from shared.auth.dependencies import CurrentUser


def make_user(
    *,
    agency_id: str | None = "agency-1",
    roles: list[str] | None = None,
    permissions: list[str] | None = None,
) -> CurrentUser:
    return CurrentUser(
        id="user-1",
        roles=roles or [],
        permissions=permissions or ["identity:search"],
        agency_id=agency_id,
        token_jti="token-1",
    )


def test_fuzzy_name_query_supports_name_order_and_agency_scope():
    filters = IdentitySearchFilters(
        nom="Diop",
        prenom="Aminata",
        date_naissance=date(1990, 5, 14),
    )
    statement = build_identity_search_statement(filters, make_user())
    sql = str(statement.compile(
        dialect=postgresql.dialect(),
        compile_kwargs={"literal_binds": True},
    ))

    assert "similarity(" in sql
    assert "%" in sql
    assert "citizens_read.agency_id" in sql
    assert "citizens_read.date_of_birth" in sql
    assert sql.count("citizens_read.first_name") >= 2
    assert "citizens_read.last_name" in sql
    assert "0.5" in sql
    assert "0.4" in sql


def test_exact_nif_and_nin_are_ranked_as_exact_identifiers():
    filters = IdentitySearchFilters(nif="  nif-123456  ", nin=" sn-123456 ")
    statement = build_identity_search_statement(filters, make_user())
    sql = str(statement.compile(
        dialect=postgresql.dialect(),
        compile_kwargs={"literal_binds": True},
    ))

    assert filters.nif == "NIF-123456"
    assert filters.nin == "SN-123456"
    assert "citizens_read.nif =" in sql
    assert "citizens_read.national_id =" in sql
    assert "CASE" in sql
    assert "1.0" in sql


def test_super_admin_search_is_not_limited_to_one_agency():
    statement = build_identity_search_statement(
        IdentitySearchFilters(nin="SN12345678"),
        make_user(agency_id=None, roles=["SUPER_ADMIN"]),
    )
    sql = str(statement.compile(
        dialect=postgresql.dialect(),
        compile_kwargs={"literal_binds": True},
    ))

    assert "citizens_read.agency_id" not in sql


def test_search_requires_an_identity_criterion():
    with pytest.raises(ValidationError):
        IdentitySearchFilters(date_naissance=date(1990, 1, 1))


@pytest.mark.asyncio
async def test_institutional_aggregation_obeys_source_permissions():
    class DcpjProvider:
        async def fetch_citizen_record(self, identity_id: str) -> dict:
            return {"identity_id": identity_id, "case_status": "restricted"}

    request = SimpleNamespace(
        app=SimpleNamespace(
            state=SimpleNamespace(identity_search_providers={"DCPJ": DcpjProvider()})
        )
    )
    user = make_user(permissions=["identity:search", "institution:dcpj:identity:read"])

    results = await _aggregate_institutional_records(request, user, "identity-1")

    by_source = {result.source: result for result in results}
    assert by_source["DCPJ"].status == "AVAILABLE"
    assert by_source["DCPJ"].data == {
        "identity_id": "identity-1",
        "case_status": "restricted",
    }
    assert by_source["MCI"].status == "FORBIDDEN"
    assert by_source["OAVCT"].status == "FORBIDDEN"