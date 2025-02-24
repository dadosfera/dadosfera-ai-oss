import { usePipelinesApi } from "@/api/pipelines/usePipelinesApi";
import { IconButton } from "@/components/common/IconButton";
import { TabLabel, TabPanel, Tabs } from "@/components/common/Tabs";
import {
  DataTable,
  DataTableColumn,
  DataTableRow,
} from "@/components/DataTable";
import { EnvVarList } from "@/components/EnvVarList";
import { useGlobalContext } from "@/contexts/GlobalContext";
import { useProjectsContext } from "@/contexts/ProjectsContext";
import { useSessionsContext } from "@/contexts/SessionsContext";
import { useCustomRoute } from "@/hooks/useCustomRoute";
import { useRegainBrowserTabFocus } from "@/hooks/useFocusBrowserTab";
import { useSendAnalyticEvent } from "@/hooks/useSendAnalyticEvent";
import {
  PipelineSettingsTab,
  usePipelineCanvasContext,
} from "@/pipeline-view/contexts/PipelineCanvasContext";
import type {
  PipelineJson,
  Service,
  TViewPropsWithRequiredQueryArgs,
} from "@/types";
import {
  envVariablesArrayToDict,
  isValidEnvironmentVariableName,
  validatePipeline,
} from "@/utils/webserver-utils";
import AddIcon from "@mui/icons-material/Add";
import DeleteIcon from "@mui/icons-material/Delete";
import ListIcon from "@mui/icons-material/List";
import MiscellaneousServicesIcon from "@mui/icons-material/MiscellaneousServices";
import SaveIcon from "@mui/icons-material/Save";
import ViewComfyIcon from "@mui/icons-material/ViewComfy";
import Alert from "@mui/material/Alert";
import AlertTitle from "@mui/material/AlertTitle";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import LinearProgress from "@mui/material/LinearProgress";
import Stack from "@mui/material/Stack";
import { styled } from "@mui/material/styles";
import Tab from "@mui/material/Tab";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import { Link } from "@orchest/design-system";
import { fetcher, hasValue, HEADER } from "@orchest/lib-utils";
import React, { useEffect } from "react";
import { usePipelineDataContext } from "../pipeline-view/contexts/PipelineDataContext";
import { generatePipelineJsonForSaving, instantiateNewService } from "./common";
import { GenerateParametersDialog } from "./GenerateParametersDialog";
import { usePipelineSettings } from "./hooks/usePipelineSettings";
import { useReadPipelineSchemaFiles } from "./hooks/useReadPipelineSchemaFiles";
import { PipelineParameters } from "./PipelineParameters";
import ServiceForm from "./ServiceForm";
import { ServiceTemplatesDialog } from "./ServiceTemplatesDialog";

export const ParameterDocs = () => {
  return (
    <Alert severity="info">
      <>
        <Link
          target="_blank"
          href="https://orchest.readthedocs.io/en/stable/fundamentals/pipelines.html#parametrizing-pipelines"
        >
          Learn more
        </Link>
        {" about parametrizing your pipelines and steps."}
      </>
    </Alert>
  );
};

const CustomTabPanel = styled(TabPanel)(({ theme }) => ({
  padding: theme.spacing(4, 0),
}));

const scopeMap = {
  interactive: "Interactive sessions",
  noninteractive: "Job sessions",
};

const isValidMemorySize = (value: string) =>
  value.match(/^(\d+(\.\d+)?\s*(KB|MB|GB))$/);

export type IPipelineSettingsView = TViewPropsWithRequiredQueryArgs<
  "pipeline_uuid" | "project_uuid"
>;

const tabMapping: Record<PipelineSettingsTab, number> = {
  configuration: 0,
  "environment-variables": 1,
  services: 2,
};

const tabs: {
  id: PipelineSettingsTab;
  label: string;
  icon: React.ReactNode;
}[] = [
  {
    id: "configuration",
    label: "Configuration",
    icon: <ListIcon />,
  },
  {
    id: "environment-variables",
    label: "Environment variables",
    icon: <ViewComfyIcon />,
  },
  {
    id: "services",
    label: "Data Apps",
    icon: <MiscellaneousServicesIcon />,
  },
];

export const PipelineSettingsView: React.FC = () => {
  // global states
  const { dispatch } = useProjectsContext();
  const {
    state: { hasUnsavedChanges },
    setAlert,
    setConfirm,
    setAsSaved,
  } = useGlobalContext();

  useSendAnalyticEvent("view:loaded", { name: "/pipeline-settings" });

  // data from route
  const {
    projectUuid,
    pipelineUuid,
    jobUuid,
    runUuid,
    snapshotUuid,
    isReadOnly: isReadOnlyFromQueryString,
  } = useCustomRoute();

  const { pipelineJson, setPipelineJson } = usePipelineDataContext();
  const { getSession } = useSessionsContext();

  const isRunningOnSnapshot =
    hasValue(jobUuid) && hasValue(runUuid || snapshotUuid);
  const isReadOnly = isRunningOnSnapshot || isReadOnlyFromQueryString;

  const hasRegainedFocus = useRegainBrowserTabFocus();
  // Fetching data
  const {
    projectEnvVariables,
    envVariables,
    setEnvVariables,
    pipelinePath,
    setPipelinePath,
    services,
    setServices,
    settings,
    setSettings,
    pipelineName,
    setPipelineName,
    pipelineParameters,
    setPipelineParameters,
  } = usePipelineSettings({
    projectUuid,
    pipelineUuid,
    jobUuid,
    runUuid,
    snapshotUuid,
    hasRegainedFocus,
  });

  const { parameterSchema, parameterUiSchema } = useReadPipelineSchemaFiles({
    projectUuid,
    pipelineUuid,
    jobUuid,
    runUuid,
    pipelinePath,
  });

  const fetchPipelines = usePipelinesApi((api) => api.fetchForProject);

  const allServiceNames = React.useMemo(() => {
    return new Set(Object.values(services || {}).map((s) => s.name));
  }, [services]);

  const { fullscreenTab, setFullscreenTab } = usePipelineCanvasContext();

  const [tabIndex, setTabIndex] = React.useState<number>(
    hasValue(fullscreenTab) ? tabMapping[fullscreenTab] : 0
  );

  const [servicesChanged, setServicesChanged] = React.useState(false);
  const [envVarsChanged, setEnvVarsChanged] = React.useState(false);

  const session = getSession(pipelineUuid);
  if (!session && !hasUnsavedChanges && (servicesChanged || envVarsChanged)) {
    setServicesChanged(false);
    setEnvVarsChanged(false);
  }

  const hasLoaded =
    pipelineJson &&
    envVariables &&
    (isReadOnly || hasValue(projectEnvVariables));

  const [
    isGenerateParametersDialogOpen,
    setIsGenerateParametersDialogOpen,
  ] = React.useState(false);

  const showGenerateParametersDialog = () => {
    setIsGenerateParametersDialogOpen(true);
  };
  const closeGenerateParametersDialog = () => {
    setIsGenerateParametersDialogOpen(false);
  };

  // Service['order'] acts as the serial number of a service
  const onChangeService = React.useCallback(
    (order: string, service: Service) => {
      setServices((current) => {
        return { ...current, [order]: service };
      });

      setServicesChanged(true);
      setAsSaved(false);
    },
    [setServices, setAsSaved]
  );

  const deleteService = async (serviceUuid: string) => {
    setServices((current) => {
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      const { [serviceUuid]: serviceToRemove, ...remainder } = current || {};
      return remainder;
    });

    setServicesChanged(true);
    setAsSaved(false);
    return true;
  };

  const onSelectTab = (
    e: React.SyntheticEvent<Element, Event>,
    index: number
  ) => {
    const { id } = tabs[index];
    setTabIndex(index);
    setFullscreenTab(id);
  };

  const validateServiceEnvironmentVariables = (pipeline: PipelineJson) => {
    for (let serviceName in pipeline.services) {
      let service = pipeline.services[serviceName];

      if (service.env_variables === undefined) {
        continue;
      }

      for (let envVariableName of Object.keys(service.env_variables)) {
        if (!isValidEnvironmentVariableName(envVariableName)) {
          setAlert(
            "Error",
            `Invalid environment variable name: "${envVariableName}" in service "${service.name}".`
          );
          return false;
        }
      }
    }
    return true;
  };

  const saveGeneralForm = async () => {
    if (!pipelineUuid || !pipelineJson) return;
    // do not mutate the original pipelineJson
    // put all mutations together for saving
    const updatedPipelineJson = generatePipelineJsonForSaving({
      pipelineJson,
      pipelineParameters,
      pipelineName,
      services,
      settings,
    });

    let validationResult = validatePipeline(updatedPipelineJson);
    if (!validationResult.valid) {
      setAlert("Error", validationResult.errors[0]);
      return;
    }

    // Validate environment variables of services
    if (!validateServiceEnvironmentVariables(updatedPipelineJson)) {
      return;
    }

    let envVariablesObj = envVariablesArrayToDict(envVariables);
    // Do not go through if env variables are not correctly defined.
    if (envVariablesObj.status === "rejected") {
      setAlert("Error", envVariablesObj.error);
      setTabIndex(1);
      return;
    }

    // Validate pipeline level environment variables
    for (let envVariableName of Object.keys(envVariablesObj.value)) {
      if (!isValidEnvironmentVariableName(envVariableName)) {
        setAlert(
          "Error",
          `Invalid environment variable name: "${envVariableName}".`
        );
        setTabIndex(1);
        return;
      }
    }

    let formData = new FormData();
    formData.append("pipeline_json", JSON.stringify(updatedPipelineJson));

    const [pipelineJsonChanges, pipelineChanges] = await Promise.allSettled([
      fetcher<{ success: boolean; reason?: string; message?: string }>(
        `/async/pipelines/json/${projectUuid}/${pipelineUuid}`,
        { method: "PUT", body: formData }
      ),

      fetcher<{ success: boolean; reason?: string; message?: string }>(
        `/async/pipelines/${projectUuid}/${pipelineUuid}`,
        {
          method: "PUT",
          headers: HEADER.JSON,
          body: JSON.stringify({
            // `env_variables` can be saved anytime, but
            // `path` cannot be changed when there is an active session
            // JSON.strigify will remove the `undefined` value, so path won't be saved as undefined
            env_variables: envVariablesObj.value,
            path: !session ? pipelinePath : undefined,
          }),
        }
      ),
    ]);

    const errorMessages = [
      pipelineJsonChanges.status === "rejected"
        ? "pipeline definition or Notebook JSON"
        : undefined,
      pipelineChanges.status === "rejected"
        ? "environment variables"
        : undefined,
    ].filter((value) => value);

    if (errorMessages.length > 0) {
      setAlert("Error", `Could not save ${errorMessages.join(" and ")}`);
    }

    // Re-fetch all pipelines to update the zustand store.
    await fetchPipelines(projectUuid).catch();

    setPipelineJson(updatedPipelineJson);
    setAsSaved(errorMessages.length === 0);
  };

  type ServiceRow = {
    name: string;
    scope: string;
    exposed: React.ReactNode;
    authenticationRequired: React.ReactNode;
    remove: string;
  };

  const columns: DataTableColumn<ServiceRow>[] = [
    { id: "name", label: "Service" },
    { id: "scope", label: "Scope" },
    { id: "exposed", label: "Exposed" },
    { id: "authenticationRequired", label: "Authentication required" },
    {
      id: "remove",
      label: "Delete",
      render: function ServiceDeleteButton(row) {
        return (
          <IconButton
            title="Delete"
            disabled={isReadOnly}
            onClick={(e) => {
              e.preventDefault();
              e.stopPropagation();

              setConfirm(
                "Warning",
                `Are you sure you want to delete the service: ${row.name}?`,
                async (resolve) => {
                  deleteService(row.id)
                    .then(() => {
                      resolve(true);
                    })
                    .catch((error) => {
                      setAlert("Error", `Unable to delete service ${row.name}`);
                      console.error(error);
                      resolve(false);
                    });
                  return true;
                }
              );
            }}
          >
            <DeleteIcon />
          </IconButton>
        );
      },
    },
  ];

  const serviceRows = React.useMemo<DataTableRow<ServiceRow>[]>(() => {
    if (!services) return [];
    const sortedServices = Object.entries(services).sort((a, b) => {
      if (hasValue(a[1].order) && hasValue(b[1].order)) {
        return a[1].order - b[1].order;
      } else {
        return 0;
      }
    });

    return sortedServices.map(([order, service]) => {
      return {
        id: order,
        name: service.name,
        scope: service.scope
          .map((scopeAsString) => scopeMap[scopeAsString])
          .join(", "),
        exposed: service.exposed ? "Yes" : "No",
        authenticationRequired: service.requires_authentication ? "Yes" : "No",
        remove: order,
        details: (
          <ServiceForm
            key={order}
            service={service}
            services={services}
            disabled={isReadOnly}
            updateService={(updated) => onChangeService(order, updated)}
          />
        ),
      };
    });
  }, [services, isReadOnly, onChangeService]);

  const isMemorySizeValid = isValidMemorySize(
    settings?.data_passing_memory_size || ""
  );

  const [isProcess, setIsProcess] = React.useState(false);
  useEffect(() => {
    const host = window.location.host;
    setIsProcess(["app-process"].some((v) => host.includes(v)));
  }, []);
  let newTabs = [...tabs];
  if (isProcess) newTabs = newTabs.filter((tab) => tab.id !== "services");

  return (
    <Box sx={{ padding: (theme) => theme.spacing(2) }}>
      {isGenerateParametersDialogOpen && projectUuid && pipelineUuid && (
        <GenerateParametersDialog
          pipelinePath={pipelinePath}
          isOpen={isGenerateParametersDialogOpen}
          onClose={closeGenerateParametersDialog}
          projectUuid={projectUuid}
          pipelineUuid={pipelineUuid}
          jobUuid={jobUuid}
          runUuid={runUuid}
        />
      )}
      <div>
        {!hasLoaded ? (
          <LinearProgress />
        ) : (
          <div>
            <Tabs
              value={tabIndex}
              onChange={onSelectTab}
              label="View pipeline settings"
              data-test-id="pipeline-settings"
            >
              {newTabs.map((tab) => (
                <Tab
                  key={tab.id}
                  id={tab.id}
                  label={<TabLabel icon={tab.icon}>{tab.label}</TabLabel>}
                  aria-controls={tab.id}
                  data-test-id={`${tab.id}-tab`}
                />
              ))}
            </Tabs>

            <div>
              <CustomTabPanel value={tabIndex} index={0} name="configuration">
                <div className="configuration">
                  <form
                    onSubmit={(e) => {
                      e.preventDefault();
                    }}
                  >
                    <div className="columns">
                      <div className="column">
                        <h3>Name</h3>
                      </div>
                      <div className="column">
                        <TextField
                          value={pipelineName}
                          margin="normal"
                          onChange={(e) => setPipelineName(e.target.value)}
                          label="Pipeline name"
                          InputLabelProps={{ shrink: true }}
                          disabled={isReadOnly}
                          fullWidth
                          data-test-id="pipeline-settings-configuration-pipeline-name"
                        />
                      </div>
                      <div className="clear"></div>
                    </div>

                    <div className="columns">
                      <div className="column">
                        <h3>Path</h3>
                      </div>
                      <div className="column">
                        <TextField
                          value={pipelinePath}
                          margin="normal"
                          onChange={(e) => setPipelinePath(e.target.value)}
                          label="Pipeline path"
                          InputLabelProps={{ shrink: true }}
                          disabled={isReadOnly || hasValue(session)}
                          helperText={
                            session
                              ? "You need to stop the session before changing pipeline path"
                              : ""
                          }
                          fullWidth
                          data-test-id="pipeline-settings-configuration-pipeline-path"
                        />
                      </div>
                      <div className="clear"></div>
                    </div>

                    <div className="columns">
                      <div className="column">
                        <h3>Pipeline parameters</h3>
                      </div>
                      <div className="column">
                        <PipelineParameters
                          initialValue={pipelineParameters}
                          isReadOnly={isReadOnly}
                          onSave={setPipelineParameters}
                          pipelinePath={pipelinePath}
                          parameterSchema={parameterSchema}
                          parameterUiSchema={parameterUiSchema}
                          showGenerateParametersDialog={
                            showGenerateParametersDialog
                          }
                        />
                        <Box sx={{ marginTop: 2 }}>
                          <ParameterDocs />
                        </Box>
                      </div>
                      <div className="clear"></div>
                    </div>

                    <div className="columns">
                      <div className="column">
                        <h3>Max steps parallelism</h3>
                      </div>
                      <div className="column">
                        <TextField
                          value={settings.max_steps_parallelism}
                          margin="normal"
                          onChange={(e) =>
                            setSettings({
                              ...settings,
                              ...{
                                max_steps_parallelism: parseInt(e.target.value),
                              },
                            })
                          }
                          label="Max steps parallelism"
                          InputLabelProps={{ shrink: true }}
                          disabled={isReadOnly}
                          fullWidth
                          type="number"
                          data-test-id="pipeline-settings-configuration-max-steps-parallelism"
                        />
                      </div>
                      <div className="clear"></div>
                    </div>
                  </form>
                </div>
              </CustomTabPanel>
              <CustomTabPanel
                value={tabIndex}
                index={1}
                name="environment-variables"
              >
                {envVarsChanged && session && (
                  <Alert severity="warning">
                    Note: changes to environment variables require a session
                    restart to take effect.
                  </Alert>
                )}
                {isReadOnly ? (
                  <EnvVarList
                    variables={envVariables}
                    readOnly
                    data-test-id="pipeline-read-only"
                  />
                ) : (
                  <Stack
                    direction="column"
                    alignItems="flex-start"
                    spacing={3}
                    sx={{ width: "60%" }}
                  >
                    <Stack spacing={2}>
                      <Typography component="h3" variant="h6">
                        Project environment variables
                      </Typography>
                      <EnvVarList
                        variables={projectEnvVariables}
                        readOnly
                        data-test-id="project-read-only"
                      />
                    </Stack>
                    <Alert severity="info">
                      Pipeline environment variables take precedence over
                      project environment variables.
                    </Alert>
                    <Stack spacing={2}>
                      <Typography component="h3" variant="h6">
                        Pipeline environment variables
                      </Typography>
                      <EnvVarList
                        variables={envVariables}
                        setValue={setEnvVariables}
                        data-test-id="pipeline"
                      />
                    </Stack>
                  </Stack>
                )}
              </CustomTabPanel>

              {!isProcess && (
                <CustomTabPanel value={tabIndex} index={2} name="services">
                  <Stack direction="column" spacing={2} alignItems="flex-start">
                    {servicesChanged && session && (
                      <Alert severity="warning">
                        Note: changes to environment variables require a session
                        restart to take effect.
                      </Alert>
                    )}
                    <DataTable<ServiceRow>
                      hideSearch
                      id="service-list"
                      columns={columns}
                      rows={serviceRows}
                    />
                    <Alert severity="info">
                      <AlertTitle>Want to start using Data Apps?</AlertTitle>
                      <Link
                        target="_blank"
                        href="https://orchest.readthedocs.io/en/stable/fundamentals/services.html"
                        rel="noopener noreferrer"
                      >
                        Learn more about how to expand your pipeline’s
                        capabilities.
                      </Link>
                    </Alert>
                    {!isReadOnly && (
                      <ServiceTemplatesDialog
                        onSelection={(template) => {
                          const newOrder =
                            parseInt(serviceRows.slice(-1)[0]?.id || "0") + 1;
                          const newService = instantiateNewService(
                            allServiceNames,
                            template,
                            newOrder
                          );
                          onChangeService(newOrder.toString(), newService);
                        }}
                      >
                        {(openServiceTemplatesDialog) => (
                          <Button
                            startIcon={<AddIcon />}
                            variant="contained"
                            color="secondary"
                            onClick={openServiceTemplatesDialog}
                            sx={{ marginTop: (theme) => theme.spacing(2) }}
                            data-test-id="pipeline-service-add"
                          >
                            Add Data App
                          </Button>
                        )}
                      </ServiceTemplatesDialog>
                    )}
                  </Stack>
                </CustomTabPanel>
              )}
            </div>

            {!isReadOnly && (
              <Button
                variant="contained"
                onClick={saveGeneralForm}
                startIcon={<SaveIcon />}
                disabled={!isMemorySizeValid}
                data-test-id="pipeline-settings-save"
              >
                {hasUnsavedChanges ? "SAVE*" : "SAVE"}
              </Button>
            )}
          </div>
        )}
      </div>
    </Box>
  );
};
