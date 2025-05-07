import { ProjectContextMenu } from "@/components/common/ProjectContextMenu";
import { useConfirm } from "@/hooks/useConfirm";
import { useNavigate } from "@/hooks/useCustomRoute";
import { Project } from "@/types";
import { MoreHorizOutlined, PlayArrow, Stop } from "@mui/icons-material";
import { LoadingButton } from "@mui/lab";
import { Box, IconButton } from "@mui/material";
import Stack from "@mui/material/Stack";
import {
  DataGrid,
  GridColDef,
  GridRenderCellParams,
  GridRowParams,
  GridRowSelectionModel,
  GridRowsProp,
} from "@mui/x-data-grid";
import React, { useCallback, useMemo } from "react";
import { useProjectActionsInLote } from "../hooks/useProjectActionsInLote";

type ProjectRow = {
  id: string;
  project: string;
  sessions: number | undefined;
  jobs: number | undefined;
  environments: number;
  actions: string;
};

export const ProjectTable = () => {
  const navigate = useNavigate();

  const {
    projects,
    loading,
    activeSessions,
    shoutdownSessions,
  } = useProjectActionsInLote();
  const hasData = projects.length > 0;

  const [menuAnchorEl, setMenuAnchorEl] = React.useState<HTMLElement>();
  const [selectedProject, setSelectedProject] = React.useState<Project>();

  const [projectIds, setProjectIds] = React.useState<string[]>([]);

  const renderCell = useCallback(
    (params: GridRenderCellParams<any, Project>) => (
      <IconButton
        aria-label="project options"
        onClick={(event) => {
          event.stopPropagation();

          setSelectedProject(params.row);
          setMenuAnchorEl(event.target as HTMLElement);
        }}
      >
        <MoreHorizOutlined />
      </IconButton>
    ),
    []
  );

  const openProject = (uuid: string) => {
    navigate({
      route: "pipeline",
      sticky: false,
      query: { projectUuid: uuid },
    });
  };

  const columns: GridColDef[] = useMemo(
    () => [
      {
        field: "project",
        headerName: "Project",
        width: 300,
      },
      {
        field: "sessions",
        headerName: "Active sessions",
        width: 150,
      },
      {
        field: "jobs",
        headerName: "Active Jobs",
        width: 150,
      },
      {
        field: "environments",
        headerName: "Environments",
        width: 150,
      },
      {
        field: "actions",
        headerName: "Actions",
        width: 150,
        renderCell,
      },
    ],
    [renderCell]
  );

  const rows: GridRowsProp<ProjectRow> = projects.map((proj) => ({
    id: proj.uuid,
    project: proj.path,
    sessions: proj.session_count,
    jobs: proj.active_job_count,
    environments: proj.environment_count,
    actions: "",
  }));

  const shutdownSessionsSelected = useConfirm(
    () => shoutdownSessions(projectIds),
    {
      title: `Shutdown ${projectIds.length} projects ?`,
      content: "All the sessions of the selected projects will be closed",
      cancelLabel: "cancel",
      confirmLabel: "shutdown",
      confirmButtonColor: "error",
    }
  );

  const activeSessionsSelected = useConfirm(() => activeSessions(projectIds), {
    title: `Active ${projectIds.length} projects ?`,
    content: "All the sessions of the selected projects will initialized",
    cancelLabel: "cancel",
    confirmLabel: "active",
    confirmButtonColor: "primary",
  });

  return (
    <Stack gap={2} alignItems="flex-start" justifyContent="flex-start">
      <Stack direction="row" spacing={1} sx={{ mb: 1 }}>
        <LoadingButton
          size="small"
          onClick={shutdownSessionsSelected}
          disabled={projectIds.length === 0}
          loading={loading}
          startIcon={<Stop />}
          loadingPosition="start"
        >
          <span>Shutdown Sessions</span>
        </LoadingButton>
        <LoadingButton
          size="small"
          onClick={activeSessionsSelected}
          disabled={projectIds.length === 0}
          loading={loading}
          startIcon={<PlayArrow />}
          loadingPosition="start"
        >
          <span>Spin-up Sessions</span>
        </LoadingButton>
      </Stack>
      <Box height={625} width="100%" display="flex" flexDirection="column">
        <DataGrid
          checkboxSelection
          columns={columns}
          rows={rows}
          loading={!hasData}
          pageSizeOptions={[5, 10, 25, { value: -1, label: "All" }]}
          initialState={{
            pagination: {
              paginationModel: {
                pageSize: 10,
                page: 0,
              },
            },
          }}
          onRowClick={(params: GridRowParams<ProjectRow>) => {
            openProject(params.row.id);
          }}
          onRowSelectionModelChange={(rows: GridRowSelectionModel) => {
            setProjectIds(rows as string[]);
          }}
        />
      </Box>

      {selectedProject && (
        <ProjectContextMenu
          anchorEl={menuAnchorEl}
          project={selectedProject}
          onClose={() => setSelectedProject(undefined)}
        />
      )}
    </Stack>
  );
};
