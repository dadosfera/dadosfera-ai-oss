"""empty message

Revision ID: 40a642486f67
Revises:
Create Date: 2020-12-18 13:21:01.461933

"""
import sqlalchemy as sa
from alembic import op

# revision identifiers, used by Alembic.
revision = "40a642486f67"
down_revision = None
branch_labels = None
depends_on = None


def upgrade():
    # ### commands auto generated by Alembic - please adjust! ###
    op.create_table(
        "background_tasks",
        sa.Column("task_uuid", sa.String(length=36), nullable=False),
        sa.Column("task_type", sa.String(length=50), nullable=True),
        sa.Column("status", sa.String(length=15), nullable=False),
        sa.Column("code", sa.String(length=15), nullable=True),
        sa.Column("result", sa.String(), nullable=True),
        sa.PrimaryKeyConstraint("task_uuid", name=op.f("pk_background_tasks")),
        sa.UniqueConstraint("task_uuid", name=op.f("uq_background_tasks_task_uuid")),
    )
    op.create_table(
        "datasources",
        sa.Column("name", sa.String(length=255), nullable=False),
        sa.Column("source_type", sa.String(length=100), nullable=False),
        sa.Column("connection_details", sa.JSON(), nullable=False),
        sa.Column(
            "created",
            sa.DateTime(),
            server_default=sa.text("timezone('utc', now())"),
            nullable=False,
        ),
        sa.PrimaryKeyConstraint("name", name=op.f("pk_datasources")),
        sa.UniqueConstraint("name", name=op.f("uq_datasources_name")),
    )
    op.create_table(
        "environments",
        sa.Column("uuid", sa.String(length=255), nullable=False),
        sa.Column("name", sa.String(length=255), nullable=False),
        sa.Column("project_uuid", sa.String(length=255), nullable=False),
        sa.Column("language", sa.String(length=255), nullable=False),
        sa.Column(
            "setup_script", sa.String(length=255), server_default="", nullable=True
        ),
        sa.Column("base_image", sa.String(length=255), nullable=False),
        sa.Column(
            "gpu_support", sa.Boolean(), server_default=sa.text("false"), nullable=False
        ),
        sa.PrimaryKeyConstraint("uuid", name=op.f("pk_environments")),
        sa.UniqueConstraint("uuid", name=op.f("uq_environments_uuid")),
    )
    op.create_table(
        "project",
        sa.Column("uuid", sa.String(length=255), nullable=False),
        sa.Column("path", sa.String(length=255), nullable=False),
        sa.PrimaryKeyConstraint("uuid", name=op.f("pk_project")),
        sa.UniqueConstraint("uuid", "path", name=op.f("uq_project_uuid_path")),
    )
    op.create_table(
        "experiments",
        sa.Column("name", sa.String(length=255), nullable=False),
        sa.Column("uuid", sa.String(length=255), nullable=False),
        sa.Column("pipeline_uuid", sa.String(length=255), nullable=False),
        sa.Column("project_uuid", sa.String(length=255), nullable=False),
        sa.Column("pipeline_name", sa.String(length=255), nullable=False),
        sa.Column("pipeline_path", sa.String(length=255), nullable=False),
        sa.Column(
            "created",
            sa.DateTime(),
            server_default=sa.text("timezone('utc', now())"),
            nullable=False,
        ),
        sa.Column("strategy_json", sa.Text(), nullable=False),
        sa.Column("draft", sa.Boolean(), nullable=True),
        sa.ForeignKeyConstraint(
            ["project_uuid"],
            ["project.uuid"],
            name=op.f("fk_experiments_project_uuid_project"),
            ondelete="CASCADE",
        ),
        sa.PrimaryKeyConstraint("uuid", name=op.f("pk_experiments")),
        sa.UniqueConstraint("uuid", name=op.f("uq_experiments_uuid")),
    )
    op.create_table(
        "pipeline",
        sa.Column("uuid", sa.String(length=255), nullable=False),
        sa.Column("project_uuid", sa.String(length=255), nullable=False),
        sa.Column("path", sa.String(length=255), nullable=False),
        sa.ForeignKeyConstraint(
            ["project_uuid"],
            ["project.uuid"],
            name=op.f("fk_pipeline_project_uuid_project"),
            ondelete="CASCADE",
        ),
        sa.PrimaryKeyConstraint("uuid", "project_uuid", name=op.f("pk_pipeline")),
        sa.UniqueConstraint(
            "uuid", "project_uuid", name=op.f("uq_pipeline_uuid_project_uuid")
        ),
    )
    op.create_table(
        "pipelineruns",
        sa.Column("uuid", sa.String(length=255), nullable=False),
        sa.Column("id", sa.Integer(), nullable=True),
        sa.Column("experiment", sa.String(length=255), nullable=True),
        sa.Column("parameter_json", sa.JSON(), nullable=False),
        sa.ForeignKeyConstraint(
            ["experiment"],
            ["experiments.uuid"],
            name=op.f("fk_pipelineruns_experiment_experiments"),
            ondelete="CASCADE",
        ),
        sa.PrimaryKeyConstraint("uuid", name=op.f("pk_pipelineruns")),
        sa.UniqueConstraint("uuid", name=op.f("uq_pipelineruns_uuid")),
    )
    # ### end Alembic commands ###


def downgrade():
    # ### commands auto generated by Alembic - please adjust! ###
    op.drop_table("pipelineruns")
    op.drop_table("pipeline")
    op.drop_table("experiments")
    op.drop_table("project")
    op.drop_table("environments")
    op.drop_table("datasources")
    op.drop_table("background_tasks")
    # ### end Alembic commands ###
