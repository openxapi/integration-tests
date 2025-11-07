0. You are going to use SDK to write integration tests, your goal is to make sure all the SDK functions work well.

1. SDK locations: WebSocket SDK at ../../../../../../binance-go/ws/options-streams and REST SDK at ../../../../../../binance-go/rest/options. You are only allowed to update the integration tests under the current working directory; you are not allowed to update the SDK code. If you find any SDK issues, please report them first.

2. You should read the SDK code first, learn what can be done using the SDK

3. For Streams, please write integration tests for each channel in dedicated go files, for each channel, you should write test cases for each request method, verify the response to make sure every field value is correct, you should also write test cases for each event handler, if you do not know how to trigger the event, please search the internet, please make sure the field values of the event you received are correct. There should be a IntegrationTestSuite for each channel, so that we can run all the channel related tests in one command.

4. You should read all the exported methods of the SDK, your integration tests should cover all the exported methods to make sure the SDK works well. For channel, your integrations tests should cover all the exported methods and event handlers for the channel. You should record the coverage progress, so that next time you can remember where to continue your work.
