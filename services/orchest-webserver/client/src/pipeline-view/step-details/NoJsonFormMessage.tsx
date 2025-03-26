import { JsonSchemaType } from "@/hooks/useOpenSchemaFile";
import AddIcon from "@mui/icons-material/Add";
import EditIcon from "@mui/icons-material/Edit";
import OpenInNewIcon from "@mui/icons-material/OpenInNew";
import Button from "@mui/material/Button";
import Link from "@mui/material/Link";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import "codemirror/mode/javascript/javascript";
import React from "react";

type DocLinkProps = {
  href: string;
  children: React.ReactNode;
};

const DocLink = ({ href, children }: DocLinkProps) => (
  <Link
    variant="body2"
    underline="hover"
    sx={{
      display: "inline-flex",
      flexDirection: "row",
      alignItems: "center",
    }}
    href={href}
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

// TODO: Add the right link
const STEP_PARAMETER_DOC_LINK = "https://orchest.readthedocs.io/en/stable/";

type NoJsonFormMessageContainerProps = {
  children: React.ReactNode;
};
const NoJsonFormMessageContainer = ({
  children,
}: NoJsonFormMessageContainerProps) => {
  return (
    <Stack
      alignItems="flex-start"
      spacing={2}
      sx={{
        marginTop: (theme) => theme.spacing(2),
        maxWidth: (theme) => theme.spacing(80),
      }}
    >
      {children}
    </Stack>
  );
};

type NoSchemaMessageProps = {
  openSchemaFile: (
    event: React.MouseEvent<Element, MouseEvent>,
    type: JsonSchemaType
  ) => void;
};

export const NoSchemaPropertiesDefined = ({
  openSchemaFile,
}: NoSchemaMessageProps) => {
  return (
    <NoJsonFormMessageContainer>
      <Typography component="span" variant="body2">
        {`No properties defined in the schema of this step file. JSON Forms let
          you render step parameters as UI inputs. `}
        <DocLink href={STEP_PARAMETER_DOC_LINK}>See docs</DocLink>
      </Typography>
      <Button
        startIcon={<EditIcon />}
        onClick={(e) => openSchemaFile(e, "schema")}
      >
        Edit schema file
      </Button>
    </NoJsonFormMessageContainer>
  );
};

export const NoSchemaFile = ({ openSchemaFile }: NoSchemaMessageProps) => {
  return (
    <NoJsonFormMessageContainer>
      <Typography component="span" variant="body2">
        {`No schema or UI schema files detected for this step. JSON Forms let you
          render step parameters as UI inputs. `}
        <DocLink href={STEP_PARAMETER_DOC_LINK}>See docs</DocLink>
      </Typography>
      <Button
        startIcon={<AddIcon />}
        onClick={(e) => openSchemaFile(e, "schema")}
      >
        New schema file
      </Button>
    </NoJsonFormMessageContainer>
  );
};
