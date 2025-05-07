import { useProjectsApi } from "@/api/projects/useProjectsApi";
import { killSession, launchSession } from "@/contexts/SessionsContext";
import { useFetchPipelines } from "@/hooks/useFetchPipelines";
import { Project } from "@/types";
import { useMemo, useState } from "react";

export type ProjectActionsInLote = {
  projects: Project[];
  loading: boolean;
  shoutdownSessions: (projects: string[]) => Promise<void>;
  activeSessions: (projects: string[]) => Promise<void>;
};

export const useProjectActionsInLote = (): ProjectActionsInLote => {
  const { pipelines } = useFetchPipelines();
  const projectMap = useProjectsApi((api) => api.projects);
  const fetchAll = useProjectsApi((api) => api.fetchAll);
  const [loading, setLoading] = useState<boolean>(false);

  const items = useMemo(() => Object.values(projectMap ?? {}), [projectMap]);

  const shoutdownSessions = async (projects: string[]): Promise<void> => {
    if (!pipelines) {
      return;
    }
    setLoading(true);
    const selectedPipes: Promise<unknown>[] = pipelines
      .filter((pipe) => projects.includes(pipe.project_uuid))
      .map((pipe) => {
        return killSession({
          pipelineUuid: pipe.uuid,
          projectUuid: pipe.project_uuid,
        });
      });

    Promise.allSettled(selectedPipes).then(async () => {
      await new Promise(resolve => setTimeout(resolve, 10000));
      await fetchAll();
      setLoading(false);
    });
  };

  const activeSessions = async (projects: string[]): Promise<void> => {
    if (!pipelines) {
      return;
    }
    setLoading(true);
    const selectedPipes: Promise<unknown>[] = pipelines
      .filter((pipe) => projects.includes(pipe.project_uuid))
      .map((pipe) => {
        return launchSession({
          pipelineUuid: pipe.uuid,
          projectUuid: pipe.project_uuid,
        });
      });

    Promise.allSettled(selectedPipes).then(async () => {
      await fetchAll();
      setLoading(false);
    });
  };

  return {
    loading,
    projects: items,
    shoutdownSessions,
    activeSessions,
  };
};
