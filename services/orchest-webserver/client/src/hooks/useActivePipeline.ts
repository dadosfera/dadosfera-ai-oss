import { usePipelinesApi } from "@/api/pipelines/usePipelinesApi";
import { useCurrentQuery } from "./useCustomRoute";

export const useActivePipeline = () => {
  const { pipelineUuid, projectUuid } = useCurrentQuery();

  const find = usePipelinesApi((api) => api.find);
  // TODO: Understand why React.useMemo is making this hook return undefined when app is refreshed
  // const pipeline = React.useMemo(() => {
  //   if (!projectUuid || !pipelineUuid) return undefined;
  //   else return find(projectUuid, pipelineUuid);
  // }, [find, projectUuid, pipelineUuid]);
  // setInterval(() => {
  //   console.log("useActivePipeline interval");
  //   console.log({ pipeline: find(projectUuid, pipelineUuid); })
  // }, 1000);
  if (!projectUuid || !pipelineUuid) return undefined;
  const pipeline = find(projectUuid, pipelineUuid);
  return pipeline;
};
