import { SvgIconProps, SvgIcon } from "@mui/material";
import React from "react";

export const DoubleBarIcon = (props: SvgIconProps & { collapsed?: boolean }) => {
  const { collapsed = false, ...rest } = props;

  return (
    <SvgIcon viewBox="0 0 24 24" {...rest}>
      <line 
        x1="6" 
        y1="6" 
        x2="6" 
        y2="20" 
        stroke="#262322" 
        strokeWidth="2"
        strokeLinecap="round"
      />

      <line
        x1="12"
        y1="10"
        x2="12"
        y2="16"
        stroke="#262322"
        strokeWidth="2"
        strokeLinecap="round"
      />
    </SvgIcon>
  );
};
