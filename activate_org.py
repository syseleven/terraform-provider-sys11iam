from sqlalchemy import create_engine, MetaData, Table
import sys
from wiremock.base import RestClient
from wiremock.constants import Config, make_headers
from wiremock.client import *
from wiremock.exceptions.api_unavailable_exception import ApiUnavailableException

WIREMOCK_BASE = "localhost:10000"
Config.base_url = "http://{}/__admin".format(WIREMOCK_BASE)

engine = create_engine("postgresql+psycopg2://glue:password@localhost:5432/iam")
org_table = Table("orgs", MetaData(), autoload_with=engine)

try:
    org_id = sys.argv[1]
    print("Activating organization with ID: {0}".format(org_id))
except IndexError:
    print("Usage: {0} org_id".format(sys.argv[0]), file=sys.stderr)
    sys.exit(1)
try:
    with engine.begin() as conn:
        result = conn.execute(
            org_table.update()
            .values(is_active=True)
            .where(org_table.c.id == org_id)
            .returning(org_table.c.id)
        )
        print("Activated organization: {0}".format(result.mappings().all()[0].get("id")))
except Exception:
    print("Failed to active organization with ID: {0}".format(org_id, file=sys.stderr))
    sys.exit(1)

mappings = [
    Mapping(
            priority=100,
            request=MappingRequest(
                method=HttpMethods.POST,
                url="/smapi2/api/ncs/organizations/%s/project.json" % org_id,
            ),
            response=MappingResponse(
                status=201,
                json_body={},  # json.loads(mock.format(project_id)),
            ),
            persistent=False,
        ),
    Mapping(
            priority=100,
            request=MappingRequest(
                method=HttpMethods.DELETE,
                url="/smapi2/api/ncs/organizations/%s" % org_id,
            ),
            response=MappingResponse(
                status=200,
                json_body={},
            ),
            persistent=False,
        )
    ]
for mapping in mappings:
    Mappings.create_mapping(mapping=mapping)
