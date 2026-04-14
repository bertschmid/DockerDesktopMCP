import React, { useCallback, useEffect, useState } from 'react';
import {
  Box,
  Button,
  Chip,
  CircularProgress,
  Divider,
  Grid,
  IconButton,
  Paper,
  Stack,
  Tooltip,
  Typography,
} from '@mui/material';
import CheckCircleOutlineIcon from '@mui/icons-material/CheckCircleOutline';
import ErrorOutlineIcon from '@mui/icons-material/ErrorOutline';
import HourglassEmptyIcon from '@mui/icons-material/HourglassEmpty';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import RefreshIcon from '@mui/icons-material/Refresh';
import { createDockerDesktopClient } from '@docker/extension-api-client';

// Docker Desktop client — communicates with the extension backend
// via Docker Desktop's internal socket proxy (no hardcoded port needed).
const ddClient = createDockerDesktopClient();

// ─── Types ────────────────────────────────────────────────────────────────────

type ServerStatus = 'checking' | 'running' | 'offline';

interface HealthResponse {
  status: string;
  server: string;
  version: string;
}

// ─── Tool Definitions ─────────────────────────────────────────────────────────

const TOOL_GROUPS: Array<{
  label: string;
  color: 'primary' | 'secondary' | 'success' | 'warning' | 'info' | 'error';
  tools: string[];
}> = [
  {
    label: 'Containers',
    color: 'primary',
    tools: [
      'docker_container_list',
      'docker_container_inspect',
      'docker_container_start',
      'docker_container_stop',
      'docker_container_restart',
      'docker_container_remove',
      'docker_container_logs',
      'docker_container_exec',
      'docker_container_stats',
      'docker_container_create',
    ],
  },
  {
    label: 'Images',
    color: 'secondary',
    tools: [
      'docker_image_list',
      'docker_image_pull',
      'docker_image_remove',
      'docker_image_inspect',
      'docker_image_tag',
      'docker_image_build',
    ],
  },
  {
    label: 'Volumes',
    color: 'success',
    tools: [
      'docker_volume_list',
      'docker_volume_create',
      'docker_volume_remove',
      'docker_volume_inspect',
    ],
  },
  {
    label: 'Networks',
    color: 'warning',
    tools: [
      'docker_network_list',
      'docker_network_create',
      'docker_network_remove',
      'docker_network_inspect',
      'docker_network_connect',
      'docker_network_disconnect',
    ],
  },
  {
    label: 'Compose',
    color: 'info',
    tools: [
      'docker_compose_up',
      'docker_compose_down',
      'docker_compose_ps',
      'docker_compose_logs',
      'docker_compose_pull',
    ],
  },
  {
    label: 'System',
    color: 'error',
    tools: [
      'docker_system_info',
      'docker_system_version',
      'docker_system_df',
      'docker_system_prune',
    ],
  },
];

const TOTAL_TOOLS = TOOL_GROUPS.reduce((sum, g) => sum + g.tools.length, 0);

// MCP endpoint — the port exposed by the backend service for AI clients.
const MCP_ENDPOINT = 'http://127.0.0.1:3282/mcp';

const CLIENT_CONFIG = JSON.stringify(
  {
    mcpServers: {
      'docker-desktop': {
        type: 'http',
        url: MCP_ENDPOINT,
      },
    },
  },
  null,
  2,
);

// ─── Status Indicator ─────────────────────────────────────────────────────────

function StatusIcon({ status }: { status: ServerStatus }) {
  if (status === 'checking') return <HourglassEmptyIcon color="disabled" />;
  if (status === 'running') return <CheckCircleOutlineIcon color="success" />;
  return <ErrorOutlineIcon color="error" />;
}

function statusLabel(status: ServerStatus): string {
  if (status === 'checking') return 'Checking server…';
  if (status === 'running') return 'MCP Server is running';
  return 'MCP Server is offline';
}

// ─── App ─────────────────────────────────────────────────────────────────────

export function App() {
  const [status, setStatus] = useState<ServerStatus>('checking');
  const [serverVersion, setServerVersion] = useState<string>('');
  const [copied, setCopied] = useState(false);

  // Health-check via Docker Desktop's internal service proxy
  const checkHealth = useCallback(async () => {
    setStatus('checking');
    try {
      const result = (await ddClient.extension.vm?.service?.get(
        '/health',
      )) as HealthResponse;
      if (result?.status === 'ok') {
        setStatus('running');
        setServerVersion(result.version ?? '');
      } else {
        setStatus('offline');
      }
    } catch (_err) {
      setStatus('offline');
    }
  }, []);

  useEffect(() => {
    checkHealth();
  }, [checkHealth]);

  const handleCopy = () => {
    navigator.clipboard.writeText(CLIENT_CONFIG).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    });
  };

  return (
    <Stack spacing={3} sx={{ p: 3 }}>
      {/* ── Page Header ───────────────────────────────────────────────── */}
      <Box>
        <Typography variant="h3">Docker Desktop MCP Server</Typography>
        <Typography variant="body1" color="text.secondary" sx={{ mt: 1 }}>
          Exposes Docker Desktop's full API to AI assistants via the{' '}
          <strong>Model Context Protocol (MCP)</strong>. Connect Claude or any
          MCP-compatible client to manage containers, images, volumes, networks,
          and Compose stacks — all from natural language.
        </Typography>
      </Box>

      <Grid container spacing={2}>
        {/* ── Server Status ─────────────────────────────────────────── */}
        <Grid item xs={12} md={6}>
          <Paper variant="outlined" sx={{ p: 2, height: '100%' }}>
            <Stack
              direction="row"
              justifyContent="space-between"
              alignItems="flex-start"
            >
              <Typography variant="subtitle1" fontWeight={600} gutterBottom>
                Server Status
              </Typography>
              <Tooltip title="Refresh status">
                <span>
                  <IconButton
                    size="small"
                    onClick={checkHealth}
                    disabled={status === 'checking'}
                  >
                    {status === 'checking' ? (
                      <CircularProgress size={16} />
                    ) : (
                      <RefreshIcon fontSize="small" />
                    )}
                  </IconButton>
                </span>
              </Tooltip>
            </Stack>

            <Stack direction="row" spacing={1.5} alignItems="center">
              <StatusIcon status={status} />
              <Box>
                <Typography variant="body1" fontWeight={500}>
                  {statusLabel(status)}
                </Typography>
                {status === 'running' && (
                  <Typography
                    variant="body2"
                    color="text.secondary"
                    sx={{ fontFamily: 'monospace' }}
                  >
                    {MCP_ENDPOINT}
                    {serverVersion && ` · v${serverVersion}`}
                  </Typography>
                )}
                {status === 'offline' && (
                  <Typography variant="body2" color="text.secondary">
                    The backend service is not responding. Check that the
                    extension container is running.
                  </Typography>
                )}
              </Box>
            </Stack>
          </Paper>
        </Grid>

        {/* ── Quick Start ───────────────────────────────────────────── */}
        <Grid item xs={12} md={6}>
          <Paper variant="outlined" sx={{ p: 2, height: '100%' }}>
            <Typography variant="subtitle1" fontWeight={600} gutterBottom>
              Quick Start
            </Typography>
            <Stack spacing={0.5}>
              {[
                'Copy the client configuration snippet below.',
                'Paste it into your AI client\'s MCP server settings.',
                'Restart the AI client.',
                'Ask: "Show me all running Docker containers"',
              ].map((step, i) => (
                <Typography key={i} variant="body2" color="text.secondary">
                  <strong>{i + 1}.</strong> {step}
                </Typography>
              ))}
            </Stack>
          </Paper>
        </Grid>
      </Grid>

      {/* ── Client Configuration ──────────────────────────────────────────── */}
      <Paper variant="outlined" sx={{ p: 2 }}>
        <Stack
          direction="row"
          justifyContent="space-between"
          alignItems="center"
          sx={{ mb: 1 }}
        >
          <Typography variant="subtitle1" fontWeight={600}>
            Client Configuration
          </Typography>
          <Tooltip title={copied ? 'Copied!' : 'Copy to clipboard'}>
            <Button
              size="small"
              variant="outlined"
              startIcon={<ContentCopyIcon />}
              onClick={handleCopy}
            >
              {copied ? 'Copied!' : 'Copy'}
            </Button>
          </Tooltip>
        </Stack>
        <Typography variant="body2" color="text.secondary" paragraph>
          Add this JSON to your AI client's MCP server configuration file
          (e.g. <code>claude_desktop_config.json</code> for Claude Desktop):
        </Typography>
        <Box
          component="pre"
          sx={{
            bgcolor: 'action.hover',
            borderRadius: 1,
            p: 2,
            m: 0,
            fontFamily: 'monospace',
            fontSize: '0.8125rem',
            overflowX: 'auto',
            whiteSpace: 'pre',
          }}
        >
          {CLIENT_CONFIG}
        </Box>
      </Paper>

      {/* ── Available Tools ───────────────────────────────────────────────── */}
      <Paper variant="outlined" sx={{ p: 2 }}>
        <Typography variant="subtitle1" fontWeight={600} gutterBottom>
          Available Tools ({TOTAL_TOOLS})
        </Typography>
        <Typography variant="body2" color="text.secondary" paragraph>
          Each tool maps to a Docker operation callable by the AI assistant.
        </Typography>

        <Stack spacing={2} divider={<Divider />}>
          {TOOL_GROUPS.map((group) => (
            <Box key={group.label}>
              <Typography
                variant="overline"
                color="text.secondary"
                sx={{ display: 'block', mb: 1 }}
              >
                {group.label} ({group.tools.length})
              </Typography>
              <Stack direction="row" flexWrap="wrap" gap={1}>
                {group.tools.map((tool) => (
                  <Chip
                    key={tool}
                    label={tool}
                    size="small"
                    variant="outlined"
                    color={group.color}
                    sx={{ fontFamily: 'monospace', fontSize: '0.6875rem' }}
                  />
                ))}
              </Stack>
            </Box>
          ))}
        </Stack>
      </Paper>
    </Stack>
  );
}
