import ArrowDropDownOutlinedIcon from "@mui/icons-material/ArrowDropDownOutlined";
import Button from "@mui/material/Button";
import React, { useEffect } from "react";
import { useInteractiveRuns } from "../hooks/useInteractiveRuns";
import { ServicesMenuComponent } from "./ServicesMenuComponent";
import { useServices } from "./useServices";

export const ServicesMenu = () => {
  const { displayStatus } = useInteractiveRuns();

  const servicesButtonRef = React.useRef<HTMLButtonElement | null>(null);
  const { anchor, services, showServices, hideServices } = useServices(
    displayStatus === "RUNNING"
  );

  const [isProcess, setIsProcess] = React.useState(false);
  useEffect(() => {
    const host = window.location.host;
    setIsProcess(["app-process"].some((v) => host.includes(v)));
  }, []);

  return (
    <>
      {!isProcess && (
        <>
          <Button
            size="small"
            id="running-services-button"
            onClick={showServices}
            endIcon={<ArrowDropDownOutlinedIcon />}
            ref={servicesButtonRef}
          >
            Data Apps
          </Button>

          <ServicesMenuComponent
            onClose={hideServices}
            anchor={anchor}
            services={services}
          />
        </>
      )}
    </>
  );
};
