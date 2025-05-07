import { ResizablePane } from "@/components/ResizablePane";
import { SxProps, Theme, IconButton } from "@mui/material";
import React from "react";
import {
  MIN_MAIN_SIDE_PANEL_WIDTH,
  useMainSidePanelWidth,
} from "./stores/useLayoutStore";
import { DoubleBarIcon } from "../common/icons/DoubleBarIcon";

const paneSx: SxProps<Theme> = {
  backgroundColor: (theme) => theme.palette.grey[100],
  borderRight: (theme) => `1px solid ${theme.borderColor}`,
  display: "flex",
  flexDirection: "column",
  height: "100%",
  position: "relative",
};

export const MainSidePanel: React.FC = ({ children }) => {
  const [mainSidePanelWidth, setMainSidePanelWidth] = useMainSidePanelWidth();
  const [collapsed, setCollapsed] = React.useState(false);
  const previousWidth = React.useRef(mainSidePanelWidth);

  const handleToggleCollapse = () => {
    if (collapsed) {
      setMainSidePanelWidth(previousWidth.current);
    } else {
      previousWidth.current = mainSidePanelWidth;
      setMainSidePanelWidth(0);
    }
    setCollapsed(!collapsed);
  };

  return (
    <div style={{ height: "100%", position: "relative", display: "flex" }}>
      {!collapsed && (
        <ResizablePane
          direction="horizontal"
          initialSize={mainSidePanelWidth}
          minWidth={MIN_MAIN_SIDE_PANEL_WIDTH}
          maxWidth={window.innerWidth / 2}
          sx={paneSx}
          onSetSize={(size) => {
            if (!collapsed) {
              setMainSidePanelWidth(size);
            }
          }}
        >
          {children}
        </ResizablePane>
      )}

      <IconButton
        onClick={handleToggleCollapse}
        size="small"
        sx={{
          position: "absolute",
          top: "50%",
          transform: "translateY(-50%)",
          left: collapsed ? 5 : `calc(${mainSidePanelWidth}px)`,
          zIndex: 10,
          opacity: 0.5,
          transition: 'opacity 0.2s ease',
          "&:hover": { backgroundColor: "#fff", opacity: 1 },
        }}
      >
        <DoubleBarIcon collapsed={collapsed} />
      </IconButton>
      </div>
  );
};
