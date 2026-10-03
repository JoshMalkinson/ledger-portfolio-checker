import { createClient } from '@connectrpc/connect'
import { createGrpcWebTransport } from '@connectrpc/connect-web'
import { PortfolioService } from './gen/portfolio/v1/portfolio_pb'

const client = createClient(PortfolioService, createGrpcWebTransport({ baseUrl: window.location.origin }))

export function checkPortfolio(csv: Uint8Array, limitBasisPoints: number, signal?: AbortSignal) {
  return client.checkPortfolio({ csvData: csv, limitBasisPoints }, { signal, timeoutMs: 15000 })
}
