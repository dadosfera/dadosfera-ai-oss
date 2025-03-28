import { Project } from "@/types";
import { memoized, MemoizePending } from "@/utils/promise";
import { create } from "zustand";
import { persist } from "zustand/middleware";
import { FetchAllParams, projectsApi, ProjectUpdateData } from "./projectsApi";

export type ProjectMap = { [uuid: string]: Project };

export type ProjectsApi = {
  /** A map of the currently available projects by UUID. */
  projects: ProjectMap | undefined;
  /** A list of the project UUIDs currently being deleted. */
  deleting: string[];
  /** Loads all available projects. */
  fetchAll: MemoizePending<(params?: FetchAllParams) => Promise<ProjectMap>>;
  /** Creates a project with the  */
  create: MemoizePending<(projectName: string) => Promise<Project>>;
  /** Updates the project with the new data. */
  update: MemoizePending<
    (uuid: string, data: ProjectUpdateData) => Promise<Project>
  >;
  /** Irreversibly deletes the project and all data associated with it. */
  delete: MemoizePending<(projectUuid: string) => Promise<void>>;
  /** Assigns a new name to a project with a given UUID. */
  rename: MemoizePending<
    (projectUuid: string, newName: string) => Promise<void>
  >;
};

export const useProjectsApi = create(
  persist<ProjectsApi>(
    (set) => {
      const reload = async (params: FetchAllParams = {}) => {
        const projects = await projectsApi.fetchAll(params).then(toProjectMap);

        set({ projects });

        return projects;
      };

      const reloadAndFind = async (uuid: string, params?: FetchAllParams) => {
        const projects = await reload(params);
        const project = projects[uuid];

        if (project) return project;

        throw new Error(`The project ${uuid} wasn't found after reload.`);
      };

      return {
        projects: undefined,
        deleting: [],
        fetchAll: memoized(reload),
        create: memoized(async (name) => {
          const uuid = await projectsApi.post(name);
          return await reloadAndFind(uuid);
        }),
        update: memoized(async (uuid, data) => {
          await projectsApi.put(uuid, data);
          return await reloadAndFind(uuid);
        }),
        rename: memoized(async (uuid: string, newName: string) => {
          await projectsApi.put(uuid, { name: newName });
          await reload();
        }),
        delete: memoized(async (projectUuid: string) => {
          set(({ deleting }) => ({ deleting: [...deleting, projectUuid] }));

          await projectsApi.delete(projectUuid);

          set(({ deleting }) => ({
            deleting: deleting.filter((uuid) => uuid !== projectUuid),
          }));

          await reload();
        }),
      };
    },
    {
      name: "orchest.projects",
      partialize: (api) => ({ ...api, projects: api.projects, deleting: [] }),
    }
  )
);

const toProjectMap = (projects: Project[]): ProjectMap =>
  Object.fromEntries(projects.map((project) => [project.uuid, project]));
