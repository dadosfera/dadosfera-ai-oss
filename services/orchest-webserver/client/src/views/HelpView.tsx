import { useOrchestConfigsApi } from "@/api/system-config/useOrchestConfigsApi";
import { PageTitle } from "@/components/common/PageTitle";
import { Layout } from "@/components/layout/Layout";
import { useOnboardingDialog } from "@/components/layout/legacy/OnboardingDialog";
import { useSendAnalyticEvent } from "@/hooks/useSendAnalyticEvent";
import { siteMap } from "@/routingConfig";
import PlayArrowIcon from "@mui/icons-material/PlayArrow";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import React from "react";

const HelpItem: React.FC<{
  link: string;
  image: string;
  imageStyle?: React.CSSProperties;
}> = ({ link, children, image, imageStyle }) => {
  return (
    <Stack
      component="a"
      direction="row"
      alignItems="center"
      spacing={3}
      href={link}
      target="_blank"
      rel="noreferrer"
      sx={{
        color: (theme) => theme.palette.text.primary,
        height: (theme) => theme.spacing(4),
      }}
    >
      <Box
        sx={{
          width: (theme) => theme.spacing(3),
          height: (theme) => theme.spacing(3),
        }}
      >
        <img src={image} style={{ width: "100%", ...imageStyle }} />
      </Box>
      <Typography>{children}</Typography>
    </Stack>
  );
};

const HelpView: React.FC = () => {
  const config = useOrchestConfigsApi((state) => state.config);

  useSendAnalyticEvent("view:loaded", { name: siteMap.help.path });
  const { setIsOnboardingDialogOpen } = useOnboardingDialog();

  const {
    ORCHEST_WEB_URLS: { readthedocs, website, slack, github },
  } = config || { ORCHEST_WEB_URLS: {} };

  return (
    <Layout>
      <div className="view-page help-list">
        <PageTitle>Looking for help, or want to know more?</PageTitle>
        <p className="push-down">
          The documentation should get you up to speed.
        </p>

        <Stack
          direction="column"
          spacing={3}
          alignItems="flex-start"
          sx={{ marginLeft: (theme) => theme.spacing(3) }}
        >
          {/* {hasValue(readthedocs) && ( */}
          <>
            {/* <HelpItem
                link={`${readthedocs}/getting_started/quickstart.html`}
                image="/image/readthedocs.png"
                imageStyle={{ width: "18px", margin: "0 auto" }}
              >
                Quickstart
              </HelpItem> */}
            <HelpItem
              link={"https://docs.dadosfera.ai/docs"}
              image="/image/readthedocs.png"
              imageStyle={{ width: "18px", margin: "0 auto" }}
            >
              Documentation
            </HelpItem>
          </>
          {/* )} */}
          {/* {hasValue(website) && (
            <HelpItem
              link={`${website}/video-tutorials`}
              image="/image/logo.svg"
              imageStyle={{ minWidth: "28px", marginLeft: "-2px" }}
            >
              Video tutorials
            </HelpItem>
          )}
          {hasValue(slack) && (
            <HelpItem link={slack} image="/image/slack.png">
              Slack
            </HelpItem>
          )}
          {hasValue(github) && (
            <HelpItem link={github} image="/image/github.png">
              GitHub
            </HelpItem>
          )}
          {hasValue(website) && (
            <HelpItem
              link={website}
              image="/image/logo.svg"
              imageStyle={{ minWidth: "28px", marginLeft: "-2px" }}
            >
              Website
            </HelpItem>
          )} */}

          <HelpItem
            link={
              "https://ajuda.dadosfera.ai/_hcms/mem/login?redirect_url=https%3A%2F%2Fajuda.dadosfera.ai%2Fportal"
            }
            image="/image/logo.svg"
            imageStyle={{ minWidth: "28px", marginLeft: "-2px" }}
          >
            Support
          </HelpItem>
        </Stack>
        <PageTitle sx={{ marginTop: (theme) => theme.spacing(3) }}>
          Introduction
        </PageTitle>
        <Button
          data-test-id="onboarding-open"
          onClick={() => setIsOnboardingDialogOpen(true)}
          startIcon={<PlayArrowIcon />}
        >
          Show onboarding
        </Button>
      </div>
    </Layout>
  );
};

export default HelpView;
