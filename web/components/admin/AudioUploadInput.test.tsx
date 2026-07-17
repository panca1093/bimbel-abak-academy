import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { AudioUploadInput } from "./AudioUploadInput";

type PresignInput = { filename: string; content_type: string };
type PresignOutput = { url: string; method: "PUT"; key: string };
type PresignFn = (input: PresignInput) => Promise<PresignOutput>;

let presignState: {
  mutateAsync: PresignFn;
  isPending: boolean;
} = {
  mutateAsync: vi.fn(),
  isPending: false,
};

vi.mock("@/lib/hooks/admin-uploads", () => ({
  usePresignAdminAudioUpload: () => presignState,
  usePresignAdminImageUpload: () => ({
    mutateAsync: vi.fn(),
    isPending: false,
  }),
}));

beforeEach(() => {
  presignState = {
    mutateAsync: vi.fn().mockResolvedValue({
      url: "https://upload.example.com/put-here",
      method: "PUT",
      key: "questions/uuid/audio.mp3",
    }),
    isPending: false,
  };
});

describe("AudioUploadInput", () => {
  it("renders input field and upload button", () => {
    render(
      <AudioUploadInput
        id="test-audio"
        value=""
        onChange={vi.fn()}
      />
    );

    const input = document.querySelector('input[id="test-audio"]') as HTMLInputElement;
    expect(input).toBeInTheDocument();

    const button = screen.getByRole("button", { name: /upload audio/i });
    expect(button).toBeInTheDocument();
  });

  it("selects and uploads a file", async () => {
    const onChange = vi.fn();

    render(
      <AudioUploadInput
        id="test-audio"
        value=""
        onChange={onChange}
      />
    );

    const fetchSpy = vi.fn().mockResolvedValue({ ok: true });
    vi.stubGlobal("fetch", fetchSpy);

    // Get the hidden file input
    const fileInput = document.querySelector(
      'input[data-testid="audio-upload-input-test-audio"]'
    ) as HTMLInputElement;

    const file = new File(["audio data"], "test.mp3", { type: "audio/mpeg" });
    fireEvent.change(fileInput, { target: { files: [file] } });

    // Wait for presign call
    await waitFor(() => {
      expect(presignState.mutateAsync).toHaveBeenCalledWith({
        filename: "test.mp3",
        content_type: "audio/mpeg",
      });
    });

    // Verify fetch was called with PUT
    await waitFor(() => {
      expect(fetchSpy).toHaveBeenCalledWith(
        "https://upload.example.com/put-here",
        expect.objectContaining({
          method: "PUT",
          body: file,
        })
      );
    });

    // Verify onChange was called with the final URL
    await waitFor(() => {
      expect(onChange).toHaveBeenCalledWith(expect.stringContaining("audio.mp3"));
    });

    vi.unstubAllGlobals();
  });

  it("displays pre-existing URL in input field", () => {
    const onChange = vi.fn();

    render(
      <AudioUploadInput
        id="test-audio"
        value="https://existing.com/audio.mp3"
        onChange={onChange}
      />
    );

    const input = screen.getByDisplayValue("https://existing.com/audio.mp3");
    expect(input).toBeInTheDocument();
  });

  it("allows typing URL directly", () => {
    const onChange = vi.fn();

    render(
      <AudioUploadInput
        id="test-audio"
        value=""
        onChange={onChange}
      />
    );

    const input = document.querySelector('input[id="test-audio"]') as HTMLInputElement;
    fireEvent.change(input, { target: { value: "https://example.com/audio.mp3" } });

    expect(onChange).toHaveBeenCalledWith("https://example.com/audio.mp3");
  });

  it("disables when loading", async () => {
    presignState.isPending = true;

    const onChange = vi.fn();

    render(
      <AudioUploadInput
        id="test-audio"
        value=""
        onChange={onChange}
      />
    );

    const input = document.querySelector('input[id="test-audio"]') as HTMLInputElement;
    const button = screen.getByRole("button", { name: /upload audio/i });

    expect(input).toBeDisabled();
    expect(button).toBeDisabled();
  });
});
