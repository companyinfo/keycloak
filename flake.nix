{
  description = "Keycloak Package Development Environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
        };
        
        # Common tools used across multiple apps
        commonTools = with pkgs; [
          gum
          go_1_24
          docker-compose
          wait4x
          jq
          curl
        ];
        
        # Define app metadata for auto-generating help
        appCategories = {
          "Testing" = {
            test = "Run unit tests";
            test-coverage = "Run tests with coverage";
            test-integration = "Run integration tests";
          };
          "Code Quality" = {
            fmt = "Format code with gofmt";
            vet = "Run go vet static analysis";
          };
          "Keycloak Management" = {
            start-keycloak = "Start Keycloak container";
            stop-keycloak = "Stop Keycloak container";
            setup-keycloak = "Full Keycloak setup for testing";
            clean-keycloak = "Stop and remove all Keycloak data";
          };
          "Development Workflows" = {
            dev = "Complete dev environment setup";
            ci = "Run all CI checks (fmt, vet, test)";
          };
          "Utilities" = {
            clean = "Remove build artifacts and coverage files";
          };
        };
        
        # Generate help text from metadata
        generateHelpText = categories:
          let
            categoryHelp = category: apps:
              ''
                gum style --foreground 212 "${category}:"
              '' + pkgs.lib.concatStringsSep "\n" (
                pkgs.lib.mapAttrsToList (name: desc:
                  ''  echo "  ${pkgs.lib.fixedWidthString 18 " " name} - ${desc}"''
                ) apps
              ) + ''
  
                echo ""
              '';
          in
            pkgs.lib.concatStringsSep "" (
              pkgs.lib.mapAttrsToList categoryHelp categories
            );
      in
      {
        devShells.default = pkgs.mkShell {
          name = "keycloak-dev-environment";
          
          buildInputs = with pkgs; [
            # Go development
            go_1_24
            gopls
            gotools
            go-tools
            
            # Testing
            gotestsum
            
            # JSON processing
            jq
            
            # Docker & Compose
            docker-compose
            
            # Shell UI
            gum
            
            # Utilities
            curl
            wait4x
            
            # Version control
            git
          ];

          shellHook = ''
            echo "Keycloak Package Development Environment"
            echo "Go version: $(go version)"
            echo ""
            echo "Available Nix apps (run with 'nix run .#<app>'):"
            echo "  test               - Run unit tests"
            echo "  test-coverage      - Run tests with coverage"
            echo "  test-integration   - Run integration tests"
            echo "  fmt                - Format code"
            echo "  vet                - Run go vet"
            echo "  start-keycloak     - Start Keycloak"
            echo "  stop-keycloak      - Stop Keycloak"
            echo "  setup-keycloak     - Setup Keycloak for testing"
            echo "  clean-keycloak     - Stop and remove Keycloak data"
            echo "  dev                - Full dev environment setup"
            echo "  ci                 - Run CI checks"
            echo "  clean              - Remove build artifacts"
            echo ""
            echo "Or use make targets directly: make help"
            echo ""
          '';
        };

        # Expose common tasks as Nix apps
        apps = {
          # Default app - shows help (auto-generated from appCategories)
          default = {
            type = "app";
            program = "${pkgs.writeShellApplication {
              name = "help";
              runtimeInputs = commonTools;
              text = ''
                gum style \
                  --border double \
                  --border-foreground 212 \
                  --padding "1 2" \
                  --margin "1" \
                  "Keycloak Package Development Environment"
                
                echo ""
                gum style --bold "Available Commands (run with 'nix run .#<command>'):"
                echo ""
                
                ${generateHelpText appCategories}
                
                gum style --foreground 240 "Tip: Enter dev shell with 'nix develop'"
                gum style --foreground 240 "     Use Make targets directly: 'make help'"
              '';
            }}/bin/help";
          };
          
          # Testing
          test = {
            type = "app";
            program = "${pkgs.writeShellApplication {
              name = "test";
              runtimeInputs = commonTools;
              text = ''
                gum log --level info "Running unit tests..."
                
                if gum spin --show-output --spinner dot --title "Running unit tests..." -- \
                  go test -v -short -race ./...; then
                  gum log --level info "Tests passed successfully"
                else
                  gum log --level error "Tests failed"
                  exit 1
                fi
              '';
            }}/bin/test";
          };

          test-coverage = {
            type = "app";
            program = "${pkgs.writeShellApplication {
              name = "test-coverage";
              runtimeInputs = commonTools;
              text = ''
                if gum spin --show-output --spinner dot --title "Running tests with coverage..." -- \
                  go test -v -short -race -coverprofile=coverage.out ./...; then
                  echo ""
                  gum log --level info "Coverage summary:"
                  go tool cover -func=coverage.out | tail -1
                  echo ""
                  gum log --level info "Generate HTML report with: go tool cover -html=coverage.out -o coverage.html"
                else
                  gum log --level error "Tests failed"
                  exit 1
                fi
              '';
            }}/bin/test-coverage";
          };

          test-integration = {
            type = "app";
            program = "${pkgs.writeShellApplication {
              name = "test-integration";
              runtimeInputs = commonTools;
              text = ''
                if [ ! -f .env ]; then
                  gum log --level error ".env file not found. Run 'nix run .#setup-keycloak' first."
                  exit 1
                fi
                set -a && . ./.env && set +a
                if gum spin --show-output --spinner dot --title "Running integration tests..." -- \
                  go test -v -race -tags=integration -coverprofile=coverage-integration.out ./...; then
                  gum log --level info "Integration tests passed successfully"
                else
                  gum log --level error "Integration tests failed"
                  exit 1
                fi
              '';
            }}/bin/test-integration";
          };

          # Code quality
          fmt = {
            type = "app";
            program = "${pkgs.writeShellApplication {
              name = "fmt";
              runtimeInputs = commonTools;
              text = ''
                if gum spin --spinner dot --title "Formatting code..." -- \
                  gofmt -s -w .; then
                  gum log --level info "Code formatted successfully"
                else
                  gum log --level error "Code formatting failed"
                  exit 1
                fi
              '';
            }}/bin/fmt";
          };

          vet = {
            type = "app";
            program = "${pkgs.writeShellApplication {
              name = "vet";
              runtimeInputs = commonTools;
              text = ''
                if gum spin --show-output --spinner dot --title "Running go vet..." -- \
                  go vet ./...; then
                  gum log --level info "Vet passed successfully"
                else
                  gum log --level error "Vet failed"
                  exit 1
                fi
              '';
            }}/bin/vet";
          };

          # Keycloak management
          start-keycloak = {
            type = "app";
            program = "${pkgs.writeShellApplication {
              name = "start-keycloak";
              runtimeInputs = commonTools;
              text = ''
                gum log --level info "Starting Keycloak..."
                docker-compose up -d
                gum log --level info "Waiting for Keycloak to be ready..."
                sleep 5
                gum log --level info "Keycloak is starting. Check status with: docker compose logs -f keycloak"
              '';
            }}/bin/start-keycloak";
          };

          stop-keycloak = {
            type = "app";
            program = "${pkgs.writeShellApplication {
              name = "stop-keycloak";
              runtimeInputs = commonTools;
              text = ''
                gum log --level info "Stopping Keycloak..."
                docker-compose down
              '';
            }}/bin/stop-keycloak";
          };

          setup-keycloak = {
            type = "app";
            program = "${pkgs.writeShellApplication {
              name = "setup-keycloak";
              runtimeInputs = commonTools;
              text = ''
                gum log --level info "Setting up Keycloak..."
                
                if gum spin --show-output --spinner dot --title "Starting Keycloak..." -- \
                  docker-compose up -d; then
                  gum log --level info "Keycloak started successfully"
                else
                  gum log --level error "Failed to start Docker Compose"
                  docker-compose logs keycloak | tail -50
                  exit 1
                fi
                
                gum log --level info "Waiting for Keycloak to be fully ready..."
                
                if gum spin --show-error --spinner dot --title "Waiting for Keycloak..." -- \
                  wait4x http http://localhost:8080/realms/master --timeout 2m --interval 2s; then
                  gum log --level info "Keycloak is ready"
                else
                  gum log --level error "Keycloak failed to start"
                  docker-compose logs keycloak | tail -50
                  exit 1
                fi
                
                if gum spin --show-output --spinner dot --title "Configuring Keycloak..." -- \
                  NON_INTERACTIVE=true ./scripts/setup-keycloak.sh; then
                  gum log --level info "Keycloak configured successfully"
                else
                  gum log --level error "Keycloak configuration failed"
                  exit 1
                fi
              '';
            }}/bin/setup-keycloak";
          };

          clean-keycloak = {
            type = "app";
            program = "${pkgs.writeShellApplication {
              name = "clean-keycloak";
              runtimeInputs = commonTools;
              text = ''
                gum log --level info "Stopping and removing Keycloak data..."
                docker-compose down -v
                rm -f .env
                gum log --level info "Keycloak data removed"
              '';
            }}/bin/clean-keycloak";
          };

          # Development workflow
          dev = {
            type = "app";
            program = "${pkgs.writeShellApplication {
              name = "dev";
              runtimeInputs = commonTools;
              text = ''
                gum log --level info "Setting up development environment..."
                
                # Run setup-keycloak
                if ! docker-compose up -d; then
                  gum log --level error "Failed to start Docker Compose"
                  exit 1
                fi
                
                gum log --level info "Waiting for Keycloak to be fully ready..."
                if gum spin --show-error --spinner dot --title "Waiting for Keycloak..." -- \
                  wait4x http http://localhost:8080/realms/master --timeout 2m --interval 2s; then
                  gum log --level info "Keycloak is ready"
                else
                  gum log --level error "Keycloak failed to start"
                  exit 1
                fi
                
                gum log --level info "Configuring Keycloak..."
                if ! NON_INTERACTIVE=true ./scripts/setup-keycloak.sh; then
                  gum log --level error "Keycloak configuration failed"
                  exit 1
                fi
                
                echo ""
                gum log --level info "Development environment ready!"
                echo ""
                echo "Next steps:"
                echo "  1. Load environment: source .env"
                echo "  2. Run tests: nix run .#test-integration"
                echo "  3. Access Keycloak: http://localhost:8080 (admin/admin)"
              '';
            }}/bin/dev";
          };

          # CI simulation
          ci = {
            type = "app";
            program = "${pkgs.writeShellApplication {
              name = "ci";
              runtimeInputs = commonTools;
              text = ''
                gum log --level info "Running CI checks..."
                echo ""
                
                if gum spin --spinner dot --title "Formatting code..." -- \
                  gofmt -s -w .; then
                  gum log --level info "Code formatted successfully"
                else
                  gum log --level error "Code formatting failed"
                  exit 1
                fi
                
                if gum spin --show-output --spinner dot --title "Running go vet..." -- \
                  go vet ./...; then
                  gum log --level info "Vet passed successfully"
                else
                  gum log --level error "Vet failed"
                  exit 1
                fi
                
                if gum spin --show-output --spinner dot --title "Running tests with coverage..." -- \
                  go test -v -short -race -coverprofile=coverage.out ./...; then
                  gum log --level info "Tests completed successfully"
                else
                  gum log --level error "Tests failed"
                  exit 1
                fi
                
                echo ""
                gum log --level info "All CI checks passed"
              '';
            }}/bin/ci";
          };

          # Cleanup
          clean = {
            type = "app";
            program = "${pkgs.writeShellApplication {
              name = "clean";
              runtimeInputs = commonTools;
              text = ''
                gum log --level info "Removing build artifacts and coverage files..."
                rm -f coverage.out coverage-integration.out coverage.txt coverage.html
                gum log --level info "Build artifacts cleaned"
              '';
            }}/bin/clean";
          };
        };
      }
    );
}

