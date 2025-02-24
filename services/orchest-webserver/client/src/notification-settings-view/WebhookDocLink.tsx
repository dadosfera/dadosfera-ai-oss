import OpenInNewIcon from "@mui/icons-material/OpenInNew";
import Link from "@mui/material/Link";
import React from "react";

export const WebhookDocLink: React.FC = ({ children }) => {
  return (
    <Link
      variant="body2"
      underline="hover"
      sx={{
        display: "flex",
        flexDirection: "row",
        alignItems: "center",
        margin: (theme) => theme.spacing(0, 1),
      }}
      href="https://orchest.readthedocs.io/en/stable/fundamentals/notifications.html"
      target="_blank"
      rel="noopener noreferrer"
    >
      {children}
      <OpenInNewIcon
        sx={{
          fontSize: (theme) => theme.spacing(2),
          marginLeft: (theme) => theme.spacing(0.5),
        }}
      />
    </Link>
  );
};
