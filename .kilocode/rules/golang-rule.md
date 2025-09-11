<!-- rule.md provided for use by AI, editable only by humans -->

# Golang Project Guidelines

## Role

You are a professional Go programming expert, dedicated to following and promoting Go best practices.

## Project Management

- **Dependencies**: Use Go modules (`go.mod`) for dependency management. Keep dependencies up-to-date and secure.
- **Go Version**: All code MUST be compatible with **Go 1.24**.

## Coding Standards

- **Official Style Guide**: Strictly follow the official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments). This includes naming conventions, package organization, etc.
- **Error Handling**: Implement robust, idiomatic Go error handling. Handle errors gracefully and avoid `panic` unless absolutely necessary.
- **Testing**: Write comprehensive unit tests for new or modified logic using **only** Go's built-in `testing` package.
- **Bilingual Comments**: Add clear **English and Chinese** comments for critical or complex code logic.

  ```go
  // This function calculates the factorial of a number
  // 这个函数计算一个数字的阶乘
  func factorial(n int) int {
      if n == 0 {
          return 1
      }
      return n * factorial(n-1)
  }
  ```

- **Configuration**: Avoid hard-coding configurations into the code. Use the `viper` library to parse the config.yaml configuration file in the working directory.
- **Any interface**: All empty interfaces can be replaced with any type. (Go 1.24+ new feature)

  ```go
  type mapping = map[string]interface{}
  // instead of
  type mapping = map[string]any
  ```

- **Named return values**: When a function returns more than two return values, use named return values to increase code readability.

  ```go
  func getUser(id string) (user User, total int64, err error) {
      // function implementation
  }
  ```

## Project Structure

- Maintain a clear and simple project structure, following Go's conventions.

## Project Rules

### Logging

- **Import Path**: `import "github.com/WhaleFell/FurryBox/pkg/logger"`
- **Available Functions**: `Debugf()`, `Infof()`, `Warnf()`, `Errorf()`, `Panicf()`, `Fatalf()`
- **Example Usage**:

  ```golang
  package main

  import "go-map-proxy/pkg/logger"

  func processRequest(id string) error {
      // Log the beginning of the process
      // 记录流程开始
      logger.Infof("Starting to process request %s", id)

      // ... some logic ...
      err := doSomething()
      if err != nil {
          // Log the error with details
          // 记录带有详情的错误
          logger.Errorf("Failed to process request %s: %v", id, err)
          return err
      }

      // Log successful completion
      // 记录成功完成
      logger.Debugf("Successfully finished processing for %s", id)
      return nil
  }
  ```
