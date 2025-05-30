import { IconRedis, IconStreamlit, IconTensorBoard, IconVSCode } from "@/icons";
import { Service } from "@/types";
import DraftsOutlined from "@mui/icons-material/DraftsOutlined";
import React from "react";

export type ServiceTemplate = {
  label: string;
  icon: React.ReactNode;
  config: Omit<Service, "order">;
};

export type ServiceTemplates = Record<string, ServiceTemplate>;

export const templates: ServiceTemplates = {
  tensorboard: {
    label: "TensorBoard",
    icon: <IconTensorBoard />,
    config: {
      binds: {
        "/data": "/data",
      },
      command: "bash",
      args: "-c 'umask 002 && tensorboard --logdir /data --host 0.0.0.0'",
      exposed: true,
      requires_authentication: true,
      image: "tensorflow/tensorflow",
      name: "tensorboard",
      ports: [6006],
      scope: ["interactive"],
    },
  },
  streamlit: {
    label: "Streamlit",
    icon: <IconStreamlit />,
    config: {
      binds: {
        "/data": "/data",
        "/project-dir": "/project-dir",
      },
      args: "-c 'umask 002 && streamlit run /project-dir/src/Home.py'",
      command: "bash",
      env_variables: {
        STREAMLIT_SERVER_BASE_URL_PATH: "$BASE_PATH_PREFIX_8501",
      },
      exposed: true,
      requires_authentication: true,
      image: "orchest/streamlit",
      name: "streamlit",
      ports: [8501],
      preserve_base_path: true,
      scope: ["interactive", "noninteractive"],
    },
  },
  vscode: {
    label: "VSCode",
    icon: <IconVSCode />,
    config: {
      binds: {
        "/project-dir": "/home/coder/code-server",
      },
      command: "bash",
      args:
        "-c 'umask 002 && code-server --auth none --bind-addr 0.0.0.0:8080 /home/coder/code-server'",
      exposed: true,
      requires_authentication: true,
      image: "codercom/code-server:latest",
      name: "vscode",
      ports: [8080],
      scope: ["interactive"],
    },
  },
  redis: {
    label: "Redis",
    icon: <IconRedis />,
    config: {
      exposed: true,
      requires_authentication: true,
      image: "redis",
      name: "redis",
      ports: [6379],
      scope: ["interactive", "noninteractive"],
    },
  },
  empty: {
    label: "Create custom service",
    icon: <DraftsOutlined fontSize="small" />,
    config: {
      exposed: false,
      requires_authentication: true,
      image: "",
      name: "my-service",
      scope: ["interactive", "noninteractive"],
      ports: [8000],
    },
  },
};
