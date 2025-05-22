import { useOrchestConfigsApi } from "@/api/system-config/useOrchestConfigsApi";
import { IconButton } from "@/components/common/IconButton";
import { useNavigate } from "@/hooks/useCustomRoute";
import LogoutIcon from "@mui/icons-material/Logout";
import { Button } from "@mui/material";
import AppBar from "@mui/material/AppBar";
import Stack from "@mui/material/Stack";
import Tabs from "@mui/material/Tabs";
import Toolbar from "@mui/material/Toolbar";
import React from "react";
import { useLocation } from "react-router-dom";
import { ProjectSelector } from "../project-selector/ProjectSelector";
import { CustomTab } from "./CustomTab";
import { NavigationTabs } from "./NavigationTabs";
import { SessionStatus } from "./SessionStatus";

export const HeaderBar = () => {
  const userConfig = useOrchestConfigsApi((state) => state.userConfig);
  const location = useLocation();
  const navigate = useNavigate();

  const navigateHome = () => navigate({ route: "home", sticky: false });

  const logoutHandler = () => {
    window.location.href = "/login/clear";
  };

  const logoPath = window.location.hostname.includes("process")
    ? "image/logo-process.svg"
    : "image/logo.svg";

  const redirectDadosfera = () => {
    window.open("https://app.dadosfera.ai");
  };

  return (
    <AppBar
      position="fixed"
      color="default"
      sx={{
        zIndex: (theme) => theme.zIndex.drawer + 1,
        backgroundColor: (theme) => theme.palette.background.paper,
        boxShadow: "none",
        borderBottom: (theme) => `1px solid ${theme.borderColor}`,
        height: (theme) => theme.spacing(7),
      }}
    >
      <SessionStatus />

      <Toolbar
        variant="dense"
        sx={{ justifyContent: "space-between", paddingLeft: "0 !important" }}
      >
        <Stack direction="row">
          <Tabs
            value={location.pathname === "/" ? "home" : false}
            TabIndicatorProps={{
              sx: { backgroundColor: (theme) => theme.palette.common.black },
            }}
          >
            <CustomTab
              onClick={navigateHome}
              icon={
                <img
                  src={logoPath}
                  alt="Dadosfera AI Logo"
                  width={28}
                  height={28}
                />
              }
              label="Home"
              value="home"
            />
          </Tabs>
          <ProjectSelector />
        </Stack>

        <Stack direction="row" justifyContent="flex-end">
          <NavigationTabs />

          <Button
            title="Dadosfera"
            onClick={redirectDadosfera}
            sx={{
              color: (theme) => theme.palette.action.active,
              width: (theme) => theme.spacing(7),
              hover: {
                backgroundColor: "none",
              },
            }}
          >
            <img
              src="image/dadosfera.svg"
              alt="Dadosfera Logo"
              width={24}
              height={24}
            />
          </Button>

          {userConfig?.AUTH_ENABLED && (
            <IconButton
              title="Logout"
              tabIndex={0}
              onClick={logoutHandler}
              sx={{
                color: (theme) => theme.palette.action.active,
                width: (theme) => theme.spacing(7),
              }}
            >
              <LogoutIcon />
            </IconButton>
          )}
        </Stack>
      </Toolbar>
    </AppBar>
  );
};
