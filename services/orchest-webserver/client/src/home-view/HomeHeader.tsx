import { NavigationTabs } from "@/components/common/NavigationTabs";
import { useCurrentQuery } from "@/hooks/useCustomRoute";
import HomeOutlined from "@mui/icons-material/HomeOutlined";
import Box from "@mui/material/Box";
import Stack from "@mui/material/Stack";
import Tab from "@mui/material/Tab";
import Typography from "@mui/material/Typography";
import React from "react";
import { ImportProjectButton } from "./components/ImportProjectButton";
import { NewProjectButton } from "./components/NewProjectButton";
import { SubmitExampleButton } from "./components/SubmitExampleButton";

export type HomeHeaderProps = {
  scrolled: boolean;
};

export const HomeHeader = ({ scrolled }: HomeHeaderProps) => {
  const { tab = "projects" } = useCurrentQuery();
  const isExamplesTab = tab === "examples";

  let title = "Home";
  if (window) {
    const host = window.location.host;
    if (host.includes("app-intelligence") || host.includes("192.168.49.2")) {
      title = "Home - Dadosfera Intelligence";
    } else if (host.includes("app-process")) {
      title = "Home - Dadosfera Process";
    }
  }

  return (
    <Stack>
      <Stack
        direction="row"
        alignItems="space-between"
        justifyContent="flex-start"
        paddingX={2}
        paddingY={1}
      >
        <Stack direction="row" spacing={2} flex={1} alignItems="center">
          <HomeOutlined fontSize="large" />
          <Typography variant="h4">{title}</Typography>
        </Stack>
        <Stack
          direction="row"
          justifyContent="flex-start"
          alignItems="center"
          spacing={1}
        >
          {isExamplesTab ? (
            <SubmitExampleButton />
          ) : (
            <>
              <ImportProjectButton showSuccessDialog={true} />
              <NewProjectButton />
            </>
          )}
        </Stack>
      </Stack>

      <Box sx={{ borderBottom: scrolled ? 0 : 1, borderColor: "divider" }}>
        <NavigationTabs defaultTab="projects">
          <Tab label="Projects" value="projects" />
          {/* <Tab label="Examples" value="examples" /> */}
          <Tab label="All runs" value="all-runs" />
        </NavigationTabs>
      </Box>
    </Stack>
  );
};
