import { useProjectJobsApi } from "@/api/jobs/useProjectJobsApi";
import { RouteLink } from "@/components/RouteLink";
import { useGlobalContext } from "@/contexts/GlobalContext";
import { useActivePipeline } from "@/hooks/useActivePipeline";
import { useAsync } from "@/hooks/useAsync";
import { useCurrentQuery, useCustomRoute } from "@/hooks/useCustomRoute";
import { useProjectPipelines } from "@/hooks/useProjectPipelines";
import { siteMap } from "@/routingConfig";
import { JobData, PipelineMetaData } from "@/types";
import { getUniqueName } from "@/utils/getUniqueName";
import { queryArgs } from "@/utils/text";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import { hasValue } from "@orchest/lib-utils";
import React from "react";

type InvalidEnvironmentsErrorProps = {
  invalidPipelines: string[];
};

const InvalidEnvironmentsError = ({
  invalidPipelines,
}: InvalidEnvironmentsErrorProps) => {
  const { projectUuid } = useCustomRoute();
  const pipelines = useProjectPipelines(projectUuid);
  const { deletePromptMessage } = useGlobalContext();
  const hasMultiple = invalidPipelines.length > 0;

  return (
    <>
      <Typography sx={{ marginBottom: (theme) => theme.spacing(2) }}>
        {`Unable to create a new Job. The following Pipeline${
          hasMultiple ? "s" : ""
        } contain${
          hasMultiple ? "" : "s"
        } Steps or Services with an invalid Environment. Please make sure all Pipeline Steps and Services are assigned an
    Environment that exists in the Project.`}
      </Typography>
      <Stack direction="column">
        {invalidPipelines.map((pipelineUuid) => {
          const url = `${siteMap.pipeline.path}?${queryArgs({
            projectUuid,
            pipelineUuid,
          })}`;
          const pipeline = pipelines?.find(
            (pipeline) => pipeline.uuid === pipelineUuid
          );
          return (
            <RouteLink
              key={pipelineUuid}
              underline="none"
              to={url}
              onClick={deletePromptMessage}
            >
              {pipeline?.path}
            </RouteLink>
          );
        })}
      </Stack>
    </>
  );
};

// TODO: replace this with usePipelinesApi using zustand.
const useFirstBestPipeline = (desired: PipelineMetaData | undefined) => {
  const { projectUuid } = useCurrentQuery();
  const activePipeline = useActivePipeline();
  const projectPipelines = useProjectPipelines(projectUuid);

  return desired || activePipeline || projectPipelines?.[0];
};

export const useCreateJob = (desiredPipeline?: PipelineMetaData) => {
  const pipeline = useFirstBestPipeline(desiredPipeline);
  const { setAlert } = useGlobalContext();
  const { name, uuid: pipelineUuid } = pipeline || {};
  const jobs = useProjectJobsApi((state) => state.jobs || []);
  const post = useProjectJobsApi((state) => state.post);

  const newJobName = React.useMemo(() => {
    return getUniqueName(
      "Job",
      jobs.map((job) => job.name)
    );
  }, [jobs]);

  const { run, status } = useAsync<JobData | undefined>();

  const canCreateJob =
    status !== "PENDING" && hasValue(pipelineUuid) && hasValue(name);

  const createJob = React.useCallback(async () => {
    if (canCreateJob) {
      try {
        return await run(post(pipelineUuid, name, newJobName));
      } catch (error) {
        const invalidPipelines: string[] | undefined =
          error.body?.invalid_pipelines;
        if (!invalidPipelines) {
          setAlert("Notice", "Unable to create a new Job. Please try again.");
          return;
        }
        setAlert(
          "Notice",
          <InvalidEnvironmentsError invalidPipelines={invalidPipelines} />
        );
      }
    }
  }, [post, canCreateJob, pipelineUuid, name, newJobName, run, setAlert]);

  return {
    /** Creates the new job. */
    createJob,
    /** Whether the job can be created at the current moment in time. */
    canCreateJob,
    /** The pipeline that the job will be created for. */
    pipeline,
  };
};
