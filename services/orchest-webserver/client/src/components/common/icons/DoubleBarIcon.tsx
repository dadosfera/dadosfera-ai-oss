import { SvgIconProps, SvgIcon } from "@mui/material";
import React from "react";

export const DoubleBarIcon = (props: SvgIconProps & { collapsed?: boolean }) => {
  const { collapsed = false, ...rest } = props;

  return (
    <SvgIcon {...rest}>
      <line 
        x1="6" 
        y1="6" 
        x2="6" 
        y2="20" 
        stroke="currentColor" 
        strokeWidth="2"
      />

      <line
        x1="12"
        y1="10"
        x2="12"
        y2="16"
        stroke="currentColor"
        strokeWidth="2"
      />
    </SvgIcon>
  );
};
