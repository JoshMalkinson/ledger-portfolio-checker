# GCP deployment follow-on

The delivered and tested environment is local Windows. No GCP resources have been created and no public URL exists. Docker and gcloud were not found on PATH during development.

To deploy later:

1. Choose a GCP project, account and region with the user; establish the intended cost limits.
2. Add and locally test a multistage container: Node builds `web/dist`; Go builds a Linux binary; the runtime image contains the binary and static frontend.
3. Set `HOST=0.0.0.0`; Cloud Run supplies `PORT`. Point `STATIC_DIR` at the copied frontend assets. The Go server supports unencrypted HTTP/2 for Cloud Run ingress and native gRPC.
4. Configure end-to-end HTTP/2, restricted ingress/access as appropriate, limited instance counts and request timeouts. Decide explicitly whether the fictional demo should be publicly accessible.
5. Verify `/healthz`, a real gRPC-Web browser upload, native gRPC, invalid input and the production page. Document the actual deployment and cleanup commands after testing them against the selected project.

Do not treat a successful container build as proof of deployment. Before accepting real customer data, add the identity, security, privacy, persistence and operational controls appropriate to the real requirements.

Official reference: https://cloud.google.com/run/docs/triggering/grpc
