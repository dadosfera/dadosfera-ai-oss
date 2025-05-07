// services/orchest-webserver/client/src/views/AnalyticsView.tsx
import React, { useState } from "react";
import { Layout } from "@/components/layout/Layout";
import { PageTitle } from "@/components/common/PageTitle";
import { useSendAnalyticEvent } from "@/hooks/useSendAnalyticEvent";
import { siteMap } from "@/routingConfig";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";

const AnalyticsView: React.FC = () => {
  const [iframeError, setIframeError] = useState(false);
  useSendAnalyticEvent("view:loaded", { name: siteMap.analytics.path });
  const currentDomain = window.location.origin;
  const dataAppUrl = `${currentDomain}/app-analytics`;

  return (
    <Layout>
      <Box sx={{ padding: (theme) => theme.spacing(3), height: "100%" }}>
        <PageTitle>Analytics</PageTitle>
        <Box
          sx={{
            height: "calc(100% - 64px)",
            width: "100%",
          }}
        >
          {iframeError ? (
            <Typography color="error">
              Failed to load the Analytics Data App. Please check the URL or try again later.
            </Typography>
          ) : (
            <iframe
              src={dataAppUrl}
              title="Data App Analytics"
              style={{ border: "none", width: "100%", height: "100%" }}
              allowFullScreen
              onError={() => setIframeError(true)}
            />
          )}
        </Box>
      </Box>
    </Layout>
  );
};

export default AnalyticsView;
