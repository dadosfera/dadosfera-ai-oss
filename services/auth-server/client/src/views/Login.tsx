import {
  Box,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  Divider,
} from "@mui/material";
import Button from "@mui/material/Button";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import { fetcher } from "@orchest/lib-utils";
import React from "react";

const Login: React.FC<{
  cloud: boolean;
  cloudUrl: string;
  documentationUrl: string;
  githubUrl: string;
  videosUrl: string;
  queryArgs: string;
}> = ({ cloud, cloudUrl, queryArgs }) => {
  // TODO: proper client-side validation
  const [loginFailure, setLoginFailure] = React.useState<string>("");
  const [username, setUsername] = React.useState("");
  const [password, setPassword] = React.useState("");

  const [open, setOpen] = React.useState(false);

  const handleClickOpen = () => {
    setOpen(true);
  };

  const handleClose = () => {
    setOpen(false);
  };

  const redirectToHelpPage = () => {
    window.open(
      "https://ajuda.dadosfera.ai/_hcms/mem/login?redirect_url=https%3A%2F%2Fajuda.dadosfera.ai%2Fportal"
    );
    handleClose();
  };

  const submitLogin = async (e: React.SyntheticEvent) => {
    e.preventDefault();

    let formData = new FormData();
    formData.append("username", username);
    formData.append("password", password);

    let queryString = queryArgs ? "?" + queryArgs : "";

    try {
      const response = await fetcher<{ redirect: string }>(
        `/login/submit${queryString}`,
        { method: "POST", body: formData }
      );

      if (response.redirect) {
        window.location.href = response.redirect;
      } else {
        throw { error: "Failed to redirect." };
      }
    } catch (error) {
      if (error?.body?.error) {
        setLoginFailure(error?.body?.error);
      } else {
        setLoginFailure(
          "Unknown Error while login. Please contact administrator."
        );
      }
    }
  };
  return (
    <div className="login-holder">
      {cloud && (
        <div className="cloud-login-helper">
          <div className="text-holder">
            <img src="image/logo-white.png" width="200px" />
            <h1>You have been added to Dadosfera Intelligence Module</h1>
            <p>
              You can login with the username and password provided by the
              instance owner.
            </p>
            <p>
              To access the dashboard please <a href={cloudUrl}>login here</a>.
            </p>
          </div>
        </div>
      )}
      <div className="main-login-view">
        <div className="login-form">
          <div className="box">
            {cloud ? (
              <h2>Login</h2>
            ) : (
              <img
                src="image/dadosfera-login.svg"
                width="200px"
                className="logo"
              />
            )}
            <form method="post" onSubmit={submitLogin}>
              <Stack direction="column">
                <TextField
                  label="Username"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  name="username"
                  margin="normal"
                  autoComplete='off'
                  autoFocus
                />
                <TextField
                  label="Password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  type="password"
                  name="password"
                  margin="normal"
                  autoComplete='off'
                />
                <Button
                  type="submit"
                  variant="contained"
                  color="primary"
                  sx={{ marginY: 4 }}
                >
                  Login
                </Button>

                <Divider />
                <Box marginTop={4}>
                  <Button variant="text" onClick={handleClickOpen}>
                    Forgot your password?
                  </Button>
                </Box>
                {loginFailure && (
                  <div className="error push-up">
                    {JSON.stringify(loginFailure)}
                  </div>
                )}
              </Stack>
            </form>
          </div>
          {/* <Box
            sx={{
              marginTop: (theme) => theme.spacing(3),
              width: "100%",
              textAlign: "center",
            }}
          >
            <Link
              color="secondary"
              target="_blank"
              href={documentationUrl}
              rel="noreferrer"
              sx={{ color: "grey.600" }}
            >
              Documentation
            </Link>
            <Box component="span"> - </Box>
            <Link
              color="secondary"
              target="_blank"
              href={githubUrl}
              rel="noreferrer"
              sx={{ color: "grey.600" }}
            >
              GitHub
            </Link>
            <Box component="span"> - </Box>
            <Link
              color="secondary"
              target="_blank"
              href={videosUrl}
              rel="noreferrer"
              sx={{ color: "grey.600" }}
            >
              Video tutorials
            </Link>
          </Box> */}
        </div>
      </div>

      <Dialog
        open={open}
        onClose={handleClose}
        aria-labelledby="alert-dialog-title"
        aria-describedby="alert-dialog-description"
      >
        <DialogTitle id="alert-dialog-title">Forgot your password?</DialogTitle>
        <DialogContent>
          <DialogContentText id="alert-dialog-description">
            To recover your password, contact your system administrator. Request
            a new password or help accessing your account.
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button
            onClick={handleClose}
            autoFocus
            variant="contained"
            color="error"
          >
            close
          </Button>
          <Button onClick={redirectToHelpPage} variant="contained">
            contact admin
          </Button>
        </DialogActions>
      </Dialog>
    </div>
  );
};

export default Login;
